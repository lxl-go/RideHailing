# 开发任务清单

- [x] 创建 model/carpool/ 目录及 GORM 模型
- [x] 创建 model/carpool/request/ 入参 DTO
- [x] 创建 model/carpool/response/ 出参 DTO
- [x] 创建 service/carpool/ 业务逻辑
- [x] 创建 api/v1/carpool/ handler
- [x] 创建 router/carpool/ 路由注册
- [x] 更新 api/v1/enter.go 注册 ApiGroup
- [x] 更新 service/enter.go 注册 ServiceGroup
- [x] 更新 router/enter.go 注册 RouterGroup
- [x] 更新 initialize/gorm.go 注册 AutoMigrate
- [x] go build 编译验证无报错

## 工单01 Step 2 — 乘客端 Gateway (:8081)
- [x] 创建 gateway/internal/model/ (Trip, CarpoolOrder, Review)
- [x] 创建 gateway/internal/dto/ (trip/order/review request/reply)
- [x] 创建 gateway/internal/repo/ (DB init + CRUD)
- [x] 创建 gateway/internal/service/ (SearchTrips, GetTripDetail, CreateOrder, CancelOrder, ListOrders, GetOrderDetail, SubmitReview)
- [x] 创建 gateway/internal/handler/ (Gin handlers)
- [x] 创建 gateway/router 注册 + main.go + go.mod
- [x] go build 编译验证无报错

## 工单01 Step 3 — 司机端 Gateway (:8082)
- [x] 创建 gateway/internal/model/ (模型同乘客端)
- [x] 创建 gateway/internal/dto/ (PublishTrip, MyTrips, PendingOrders, AcceptOrder, RejectOrder, SubmitReview)
- [x] 创建 gateway/internal/repo/ (DB init + CRUD)
- [x] 创建 gateway/internal/service/ (PublishTrip, MyTrips, UpdateTripStatus, PendingOrders, AcceptOrder, RejectOrder, SubmitReview)
- [x] 创建 gateway/internal/handler/ (Gin handlers)
- [x] 创建 gateway/router 注册 + main.go + go.mod
- [x] go build 编译验证无报错
