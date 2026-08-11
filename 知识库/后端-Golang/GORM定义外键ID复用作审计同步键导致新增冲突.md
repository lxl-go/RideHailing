# GORM：定义外键 ID 复用作审计同步键导致新增冲突

- 创建日期：2026-08-05
- 最近更新：2026-08-05
- 标签：`Golang` `GORM` `主键冲突` `审计/审核流` `业务同步` `错误 1062`

## 适用场景 / 问题背景

打车管理后台存在「司机提交车辆 → 管理端审核 → 通过后同步到正式车辆表」的审核流。业务上，司机对一辆**已通过的车辆**再次「编辑」时，会视为重新提交审核，并**复用原正式车辆 ID** 生成一条新的待审核记录。

前端现象：管理端对编辑后的车辆再次点击「通过」时报错「操作失败，请稍后重试」；而「驳回」正常。

## 核心结论

`syncApprovedDriverVehicle` 的同步逻辑**先按 `driver_id + plate_no`（车牌）查找正式车辆**，若司机在编辑时改了车牌，就查不到旧车，于是走 `Create(ID = 审核记录ID)` 分支。由于该 ID 已在正式车辆表中存在，插入触发**主键冲突（Error 1062 / Duplicate entry for PRIMARY）**，事务回滚，前端只看到兜底错误提示。

修复要点：**同步时应优先按「ID + driver_id」定位正式车辆并做更新；只有按 ID 找不到时，才用 `driver_id + plate_no` 兜底兼容旧数据；两者都找不到才 `Create`**。

## 关键原理

- 审核表与正式表使用的是**同一套业务主键（ID）**——审核通过后，待审核记录的 ID 即作为正式车辆记录的 ID。
- 因此，正确、稳定的业务匹配键是 **ID**，而车牌是可变的表单字段。
- 用「可变字段（车牌）」当作存在性判断键，会在字段变化时漏掉已存在的行，进而错误地走「新增」分支。
- GORM `First` 找不到时返回 `gorm.ErrRecordNotFound`；需要先判断 `errors.Is(err, gorm.ErrRecordNotFound)`，再决定是「更新」还是「新建」。任何**非 record-not-found 错误**都应直接返回，避免吞掉真实数据库错误。

## 修复示例

原逻辑（错误）：

```go
var existing carpool.DriverVehicle
err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("driver_id = ? AND plate_no = ?", info.DriverID, info.PlateNumber).
    First(&existing).Error
if err == nil {
    // 更新
} else if errors.Is(err, gorm.ErrRecordNotFound) {
    return tx.Create(...) // 车牌一改就走到这里 -> 主键冲突
}
```

修复后（正确）：

```go
lock := tx.Clauses(clause.Locking{Strength: "UPDATE"})
err := lock.Where("id = ? AND driver_id = ?", info.ID, info.DriverID).First(&existing).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    // 兼容旧数据 / 未按 ID 同步过的情况：按车牌兜底
    err = lock.Where("driver_id = ? AND plate_no = ?", info.DriverID, info.PlateNumber).First(&existing).Error
}
if err == nil {
    return tx.Model(&carpool.DriverVehicle{}).
        Where("id = ?", existing.ID).
        Updates(updates).Error
}
if !errors.Is(err, gorm.ErrRecordNotFound) {
    return err
}
return tx.Create(...)
```

> 说明：上面两段对同一 gorm.DB 复用 `lock` 变量（带行锁从句）。若两次查询 `Where` 叠加会导致条件重复，可分别创建查询构建器或复用带锁的 `tx`。

## 常见误区 / 注意事项

- **别用易变业务字段（车牌、手机号、邮箱）做存在性/幂等判断**；优先使用稳定主键或业务不可变键。
- 新增与更新分支写在一起时，务必先区分「record-not-found」与其他真实错误，否则可能掩盖 SQL 失败。
- 事务内发生主键冲突后，整个事务回滚，上层常只显示兜底文案（本案例的「操作失败，请稍后重试」），导致排查困难；应把**原始 SQL 错误日志**作为排查入口。
- 修改了同步键后，要补一个「已通过车辆编辑后换车牌，再次审核通过应更新原正式车辆、不应插入新车辆」的回归测试，断言 `Count == 1` 且旧行被更新。

## 延伸方向

- 为正式车辆表增加 **唯一索引**（如 `uk_driver_id`）从约束层面杜绝同一司机重复建车。
- 在审核记录进入 Pending 时提前记录「origin ID」，让同步逻辑更显式。
- 将「幂等 upsert」下沉为通用 helper（严格按主键 + 业务域键插查），供多个审核模型复用。