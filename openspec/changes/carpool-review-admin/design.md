# 技术设计方案

## 架构模式
GVA Gin 直接分层（适配 GVA 现有约定）：
- `api/v1/carpool/` — handler 层
- `service/carpool/` — service 层（业务逻辑 + DB 操作）
- `model/carpool/` — model 层（GORM struct）
- `model/carpool/request/` — 入参 DTO
- `model/carpool/response/` — 出参 DTO
- `router/carpool/` — 路由注册

## 数据库表
1. `carpool_certification_audit` — 认证审核表（ID Snowflake int64）
2. `carpool_vehicle_info` — 车辆信息审核表（ID Snowflake int64）

## ID 策略
所有业务主键使用 Snowflake int64，不嵌入 GVA_MODEL（GVA_MODEL 为 uint 自增）

## 路由设计
前缀 `/carpool/review`，操作记录中间件只加在写操作路由上

## 表结构
### certification_audit
- id (int64 PK)
- user_id (int64, index)
- user_role (tinyint, 1乘客 2司机)
- cert_type (tinyint, 1身份证 2驾驶证 3行驶证)
- cert_number (varchar, AES encrypted)
- real_name (varchar)
- front_image_url (varchar)
- back_image_url (varchar)
- handheld_image_url (varchar)
- status (tinyint, 0待审核 1通过 2驳回 3补充)
- reviewer_id (int64)
- reject_reason (varchar)
- submit_count (int, default 0)
- created_at/updated_at/deleted_at

### vehicle_info
- id (int64 PK)
- driver_id (int64, index)
- plate_number (varchar, UK)
- brand (varchar)
- model (varchar)
- color (varchar)
- year_check_date (datetime)
- insurance_expire (datetime)
- status (tinyint, 0待审核 1通过 2驳回)
- reviewer_id (int64)
- reject_reason (varchar)
- created_at/updated_at/deleted_at
