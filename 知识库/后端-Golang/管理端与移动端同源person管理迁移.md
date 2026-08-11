# 管理端与移动端同源：person 管理从 mock 表迁移到真实业务表

- 创建日期：2026-08-06
- 最近更新：2026-08-06
- 标签：GORM、多端数据同源、管理后台、表重构、事务写入

## 适用场景 / 问题背景

教学项目的管理端（admin-platform）人员管理此前读写 `person_profile`/`person_role` 两张「管理端专用表」，
与移动端实际使用的 `auth_user`/`driver_profile`/`passenger_profile`/`driver_vehicle`/`driver_certification` 完全割裂：
管理端看到的是种子假数据，移动端注册的真实用户管理端看不到，两侧数据无法互通。

改造目标：管理端人员列表、详情、统计、增删改、导入全部直接读写移动端真实表，保证「同源」。

## 核心结论

- 数据源统一：以 `auth_user` 为主表（存 `principal`=手机号、`role`=driver/passenger/staff、`status`），
  司机/乘客资料按角色分表取（`driver_profile`、`passenger_profile`），车牌与证件再关联 `driver_vehicle`、`driver_certification`。
- 对外 DTO 不变：`PersonDTO`/`PersonStats`/导入返回结构保持兼容，管理端接口与前端无需改动。
- 删除假数据写入：`SeedPersonDefaults` 改为空实现，不再向 `person_profile` 注入种子。
- 状态两套语义做映射：管理端 `enabled/disabled/deleted` ↔ `auth_user.status`（1/0）；司机禁用时联动 `service_status=3`。

## 关键原理

### 状态与角色映射

```go
// 管理端语义 → 真表列
func statusToAuth(s string) int { if s=="disabled"||s=="deleted" { return 0 }; return 1 }
// 角色码归一化
func roleFromCode(code string) string {
    // carpool_driver/shuttle_driver/pickup_driver → driver
}
```

- `auth_user.role` 只存 `driver/passenger/staff` 三值，管理端传入的 `carpool_driver` 等别名先归一化再落库。
- 司机状态列 `service_status`（1=离线 2=在线 3=停用）与账号状态 `status` 分离，禁用账号时联动停用司机。

### 列表 + 关键词搜索的正确性

多表关键词（手机号、姓名、昵称）不能破坏分页/计数，做法是「先聚合命中 ID 集合，再回到主表 WHERE IN」：

```go
func personIdsByKeyword(ctx, kw string) []int64 {
    // auth_user.principal LIKE；driver_profile.name/phone LIKE；passenger_profile.nickname/phone LIKE
    // 三路 Pluck 后去重进 map[int64]bool，返回 ID 集合
}
```

主查询始终是 `auth_user` 单表 `Count + Limit/Offset`，关键词命中为空集时提前返回空页，
避免 IN () 语法问题。

### 详情组装（buildDTO）

- 司机：`driver_profile`（姓名/头像）+ 首条 `driver_vehicle`（车牌/车型，按 status 升序）+ 首条 `driver_certification`（证件号/身份证）。
- 乘客：`passenger_profile`（昵称/常用地址/支付偏好）。
- 空值兜底：资料表字段为空时回退到 `principal` 手机号，避免详情出现空名字。
- 敏感字段在输出层打码（`maskPhone`/`maskIDCard`），真表存原文，展示层脱敏。

### 事务写入

创建用户时「账号 + 资料」必须同生共死：

```go
global.GVA_DB.Transaction(func(tx *gorm.DB) error {
    tx.Create(&auth_user)      // 拿到自增 ID
    tx.Create(&driver_profile{ID: user.ID, ...}) // 资料表主键 = 用户 ID
})
```

`driver_profile`/`passenger_profile` 的 `id` 直接复用 `auth_user.id`（非自增），保证一对一的键一致。

### 导入的两段式

`PreviewImport`（只校验不落库）与 `CommitImport`（校验后批量创建）共用 `previewImport(ctx, rows, commit)`；
逐行校验失败收集到 `[]PersonImportError{BatchNo, RowNo, Field, Message}`，`BatchNo = PERSON-IMP-<unixnano>`。
错误查询接口 `ListImportErrors` 走 `person_import_error` 表按 `batch_no` 分页。

## 示例

```go
type authUserRow struct {
    ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
    Principal string    `gorm:"column:principal"`
    Role      string    `gorm:"column:role"`
    Status    int       `gorm:"column:status"`
}
func (authUserRow) TableName() string { return "auth_user" }
```

本地 service 内定义镜像真实表的结构体（表名映射 `TableName()`），
避免引用移动端模型包引入跨服务耦合，代价是字段要与建表脚本保持同步。

## 常见误区 / 注意事项

- 别把「管理端字段名」直接当「真表列名」：`PersonProfile`（DTO）里的
  `PaymentPreference`/`RegisterDate`/`DriverLicenseNo` 与假表列一致，但真表是 `passenger_profile.payment_preference` 等，需显式映射。
- `role` 归一化必须用别名映射，否则移动端 `role IN (driver)` 查不到 `carpool_driver`。
- 多表关键词搜索若先查页再过滤会破坏分页计数；先过滤 ID 再分页才是正确顺序。
- 禁用/删除只做 `status` 置 0（软删），真表无 `deleted_at` 列时不要用 Delete，避免丢行。
- 废弃假数据后，`person_profile`/`person_role` 表结构仍保留（模型文件未删），但任何 service 代码不再写入。

## 延伸方向

- 列表统计的活跃度（Active）目前等于 Enabled；后续可接入订单/行程事实表统计真实活跃。
- 导入落库错误信息目前未持久化，`ListImportErrors` 会读到空表；如需要可改为写 `person_import_error`。
- 管理端与移动端共用一份表后，后续可把镜像 struct 抽取为公共包，减少字段漂移风险。
