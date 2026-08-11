# GORM AutoMigrate 收紧列为 NOT NULL 时因历史 NULL 行报 1265

## 文档信息
- 创建日期：2026-08-06
- 最近更新：2026-08-06
- 标签：`Golang`、`GORM`、`MySQL`、`AutoMigrate`、`错误1265`、`脏数据`

## 适用场景 / 问题背景
商城/管理端（Gin-Vue-Admin + GORM）启动即退出：控制台只见 `配置来源: configs\config.yaml` 与 `exit status 1`，无任何错误。原因是 zap 配置 `log-in-console: false`，日志只写文件。
从 `log/2026-08-06/error.log` 找到真实错误：

```
ALTER TABLE `carpool_trip` MODIFY COLUMN `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
Error 1265 (01000): Data truncated for column 'created_at' at row 1
```

即 `RegisterTables` → `AutoMigrate` 想把 `carpool_trip.created_at` 从**可空**收紧为 **NOT NULL**，但表中有一行 `created_at` 为 NULL → MySQL 拒绝（1265 Data truncated）。

## 核心结论
- **`Error 1265`（Data truncated）在 AutoMigrate 收紧列为 NOT NULL 时，根源是已有行包含 NULL/非法值**。不是 SQL 写错，而是存量数据不满足新约束。
- 空库时 AutoMigrate 一直正常；一旦插入「该列为 NULL」的行，下一次启动改约束就失败。
- 本项目加重的关键：**探针/临时脚本插入测试行时未带 `created_at`/`updated_at`**，而这两列在库里是可空的 → 留下 NULL 脏数据 → 阻碍后续迁移。

## 关键原理
- GORM model 里 `CreatedAt/UpdatedAt time.Time` 若无 `default` 标签，`Create` 由 GORM 自动填充 `time.Now()`；但**手动/原生 SQL/probe 插入**不经过 GORM 就不会填，若列可空则留下 NULL。
- AutoMigrate 判断列「需要收紧为 NOT NULL」时，会先 `ALTER`，MySQL 因存量 NULL（或 `0000-00-00`）无法转换而报 1265。
- 同理：把列加为 not null、或增大/收紧类型、改默认值，都可能因存量数据不符而失败。

## 修复
先定位脏数据，再补时间：
```sql
UPDATE carpool_trip SET created_at = NOW(), updated_at = NOW()
WHERE created_at IS NULL OR updated_at IS NULL;
```
（探针实测：`rows updated: 1`，随后管理端 AutoMigrate 通过、服务正常启动。）

## 常见误区 / 注意事项
- **看到 `exit status 1` 但控制台无错误**：先看 zap/flog 配置，`log-in-console: false` 时错误只写文件（`log/<日期>/error.log`）。也可临时改 `log-in-console: true` 便于调试。
- **AutoMigrate 失败报的「罪最近列」往往不是真凶**：失败可能来自任意存量行值，先查该列是否存在 NULL / `0000-00-00` 等。
- **探针/脚本写库务必带齐非空列与审计列**（created_at/updated_at），否则会污染共享表并引发类型收紧失败。
- 排查该问题用 `WHERE col IS NULL` 即可快速确认；`0000-00-00` 需用 `col='0000-00-00 00:00:00'` 或 `CAST` 排查。

## 延伸方向
- 迁移前做「存量数据校验」：对即将收紧的列先跑一次 `IS NULL`/越界统计。
- 探针类工具不要直连生产库；即使直连，也按数据库表全列模板插入。