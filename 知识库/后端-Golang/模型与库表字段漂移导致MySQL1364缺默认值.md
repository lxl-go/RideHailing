# 模型与库表字段漂移导致 MySQL 1364：缺少 publisher_id 默认值

## 文档信息
- 创建日期：2026-08-06
- 最近更新：2026-08-06
- 标签：`Golang`、`GORM`、`MySQL`、`错误1364`、`Schema漂移`

## 适用场景 / 问题背景
司机端发布行程时，时间格式等前端问题修复后，请求仍失败：
- 网关：`trip service returned status 500` → 返回「服务开小差了」。
- trip-service 数据层：`Error 1364 (HY000): Field 'publisher_id' doesn't have a default value`。
- SQL 洞察：`INSERT INTO carpool_trip (...) VALUES (...)` **没有 `publisher_id` 列**。

根因是**同一个表在多处各有一份模型定义且不一致**：建表/公共模型含 `publisher_id`，而负责 INSERT 的 trip-service 模型缺该字段 → GORM 生成 INSERT 时不含该非空列 → 数据层拒绝写入。

## 核心结论
- `Error 1364: Field 'xxx' doesn't have a default value` 通常意味着：**该列非空（not null）且无默认值，但写库代码产生的 INSERT 没带这个字段**。可能是漏收发、字段漏掉，或模型与表结构漂移。
- 本项目 `carpool_trip` 存在多份定义：
  - 公共模型（`pkg/model/carpool/trip.go`）+ 管理端模型（`admin-platform/server/model/carpool/trip.go`）：含 `publisher_id`、坐标、`origin_name/dest_name`、`departure_time`、`share_cost` 等完整字段。
  - trip-service 运行模型（`internal/data/trip.go`）：原本只有极简字段，负责 INSERT。
- 用哪份模型写库，就会用哪份字段集；写库模型缺一列就报 1364，**每补一列又会冒下一列**（连锁）。

## 关键原理
GORM 的 `Create` 默认只写入结构体上存在的字段（设了 `default` 标签的字段会省略以用库默认值）。因此：
- 结构体缺 `PublisherID` → INSERT 无 `publisher_id` 列。
- 若库表中该列 `not null` 且无默认 → MySQL 直接 1364 拒绝整条语句。

修复三件套（缺一不可，否则 SELECT 回读也丢字段）：
1. 模型 struct 补字段并带 column 标签。
2. 领域对象（biz.Trip）补字段，写入时赋值（发布行程语义上 `publisher_id = driver_id`）。
3. 模型↔领域转换函数（toDomain / toRecord）补映射。

## 修复示例
```go
// data/trip.go —— 模型补列
type tripModel struct {
    DriverID    int64 `gorm:"column:driver_id;type:bigint;not null"`
    PublisherID int64 `gorm:"column:publisher_id;type:bigint;not null"`
    // ...
}

// biz/trip.go —— 字段 + 赋值
type Trip struct { DriverID int64; PublisherID int64 /* ... */ }

func (uc *TripUsecase) PublishTrip(...) {
    trip := &Trip{ DriverID: cmd.DriverID, PublisherID: cmd.DriverID, /* ... */ }
}

// data/trip.go —— 双向映射
func toDomain(m *tripModel) *biz.Trip { /* PublisherID: m.PublisherID */ }
func toRecord(e *biz.Trip) *tripModel { /* PublisherID: e.PublisherID */ }
```

## 常见误区 / 注意事项
- **同表多份模型是事故源头**：`pkg/model`、admin 端、各 service 各持一份时，改动必须同步，否则出现本文所述 INSERT 缺列或 SELECT 缺字段。
- **1364 只针对「非空且无默认值」的列**：给该列 `DEFAULT 0` 可绕过写入问题，但会掩盖语义（发布人本应是业务必填），应优先在代码正确赋值而非改表默认值。
- **AutoMigrate 对「新增非空无默认列」无能为力**：对已有数据加列会失败；已存在的列改代码即可，不需迁移。
- **搜索技巧**：报错列全仓 grep，比对所有模型定义即可快速定位漂移位置。
- **查表必补的全量非空列**：用 `information_schema.COLUMNS WHERE IS_NULLABLE='NO' AND COLUMN_DEFAULT IS NULL` 一次性列出所有「写库必须提供」的列，避免修一列又冒一列。

## 实例：carpool_trip 修复（2026-08-06）
`SHOW COLUMNS` 显示本表含下表所列非空无默认列，trip-service 模型原本全缺，需把写库侧模型补齐并在发布时填充：

| 列 | 说明 | 移动端发布填充值 |
|---|---|---|
| publisher_id | 发布人ID | = driver_id |
| origin_name / dest_name | 点名称 | = origin / destination 文本 |
| origin_lat / origin_lng / dest_lat / dest_lng | 坐标 | 0（前端无坐标） |
| departure_time | 出发时间 | = depart_time |
| share_cost | 分摊费用 | = price |
| publisher_role / trip_type | 枚举 | 1 / 1 |

修复要点：`internal/data/trip.go` 的 `tripModel` 补全列（type 与该列在库中的 decimal/varchar 类型一致，避免 AutoMigrate 改列），`biz.Trip` 补字段，`PublishTrip` 赋值，`toDomain/toRecord` 补齐映射。之后对该库用全列 INSERT 实测通过（`OK: full insert succeeded`）。

## 延伸方向
- 可编写「模型 ↔ 库」一致性脚本，比对 `SHOW COLUMNS` 与各 Go 模型，自动告警缺列/类型漂移。
- 可考虑统一模型单一来源（如由 `pkg/model` 统一派生），从结构上消除多份定义漂移。