# Driver Vehicle Audit Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix driver vehicle audit/delete/edit flows so pending records cannot be deleted, rejected reasons are acknowledged once, approved vehicle deletion syncs to admin, and approved edits resubmit for review.

**Architecture:** Keep `driver_vehicle` as the approved operational vehicle table and `carpool_vehicle_info` as the audit table. Expose explicit vehicle item metadata to the driver app, and use `realtime_message.delivered` as the "知晓" flag for certification and vehicle rejection notifications.

**Tech Stack:** Go, GORM transactions and row locks, Kratos protobuf HTTP/gRPC generation, Vue/uni-app, Element Plus admin web.

## Global Constraints

- Pending vehicle audit rows are not deletable.
- Approved vehicle edits create a new pending audit and do not directly change the active vehicle.
- Approved vehicle deletion marks the formal vehicle inactive and syncs the latest admin audit row to "司机已删除".
- Rejected certification and vehicle notices require a "知晓" action and do not show again after acknowledgement.
- User-facing errors and page copy must be Chinese UTF-8.

---

### Task 1: Driver Service Vehicle Rules

**Files:**
- Modify: `services/driver-service/internal/biz/driver_test.go`
- Modify: `services/driver-service/internal/biz/driver.go`
- Modify: `services/driver-service/internal/biz/repo.go`
- Modify: `services/driver-service/internal/data/driver.go`

**Interfaces:**
- Produces: `DriverVehicle.Source`, `AuditID`, `ReviewStatus`, `RejectReason`, `CanEdit`, `CanDelete`
- Produces: `DriverRepo.MarkVehicleAuditDriverDeleted(ctx, driverID, plateNo string) error`

- [ ] Write failing tests for pending audit list metadata, approved edit resubmission, and approved delete audit sync.
- [ ] Run `go test ./internal/biz -run Vehicle` in `services/driver-service` and verify failures.
- [ ] Implement domain metadata and repository methods with transactions/locks.
- [ ] Re-run the focused tests and verify pass.

### Task 2: Driver Service Messages And Proto

**Files:**
- Modify: `services/driver-service/api/driver/v1/driver.proto`
- Regenerate: `services/driver-service/api/driver/v1/driver*.pb.go`
- Modify: `services/driver-service/internal/service/driver.go`
- Modify: `services/driver-service/internal/data/driver.go`
- Modify: `services/gateway-service/internal/server/http.go`
- Modify: `services/gateway-service/internal/service/driver.go`
- Modify: `services/gateway-service/internal/biz/driver.go`
- Modify: `services/gateway-service/internal/data/driver_client.go`

**Interfaces:**
- Produces: `GET /carpool/drivers/messages`
- Produces: `POST /carpool/drivers/messages/{id}/ack`

- [ ] Write failing service tests for listing undelivered audit messages and acknowledging them.
- [ ] Add protobuf messages/RPCs and regenerate generated files through `make api`.
- [ ] Wire gateway routes to driver service.
- [ ] Re-run focused driver and gateway tests.

### Task 3: Admin Review Status Sync

**Files:**
- Modify: `admin-platform/server/model/carpool/vehicle_info.go`
- Modify: `admin-platform/server/service/carpool/review_test.go`
- Modify: `admin-platform/server/service/carpool/review.go`
- Modify: `admin-platform/web/src/view/admin/workorder01/vehicle-review.vue`

**Interfaces:**
- Produces: vehicle review status `3 = 司机已删除`

- [ ] Write failing admin review test for driver-deleted status display/sync.
- [ ] Implement status constant and Chinese titles.
- [ ] Update admin web status labels.
- [ ] Run focused admin server tests.

### Task 4: Driver App UX

**Files:**
- Modify: `apps/driver-uni-app/uni-app/src/api/vehicle.js`
- Create: `apps/driver-uni-app/uni-app/src/api/message.js`
- Modify: `apps/driver-uni-app/uni-app/src/pages/vehicle/vehicle.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/certification/certification.vue`
- Modify: `apps/driver-uni-app/uni-app/src/utils/request.js`

**Interfaces:**
- Consumes: vehicle metadata fields and message list/ack endpoints.

- [ ] Hide delete for pending/rejected audit cards and show delete only for approved formal vehicles.
- [ ] Show rejection reason in a modal with a "知晓" button for vehicle and certification messages.
- [ ] Use Chinese fallbacks for 404 and request errors.
- [ ] Run uni-app build verification.
