# RideHailing Frontend Batch 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first usable frontend batch from `docs/全量工单前端页面生成方案.md`: workorder 01 real API wiring plus workorder 02~11 page and API placeholders for admin, passenger, and driver.

**Architecture:** Reuse the existing GVA admin layout and Vant mobile layouts. Admin pages are metadata-driven but rendered inside a normal Vue page with Element Plus tables, forms, drawers, and dialogs; mobile pages use Vant only and keep URL state in hash query. Real workorder 01 Gateway/Admin APIs are separated from placeholder APIs so later backend work can replace one file at a time.

**Tech Stack:** Vue 3, Vue Router, Vite, Element Plus in admin-platform/web, Vant in passenger-platform/web and driver-platform/web, Axios-compatible API wrappers.

## Global Constraints

-管理端页面必须复用 GVA 页面范式：`gva-table-box`、`gva-btn-list`、`el-table`、`el-drawer/el-dialog`、`el-pagination`。
-乘客端/司机端页面必须复用 Vant：`van-cell`、`van-field`、`van-list`、`van-popup`、`van-dialog`、`van-calendar`、`van-uploader`。
-工单01三端真实接口可联调：管理端 `/carpool/review/*`，乘客端 Gateway `http://localhost:8081`，司机端 Gateway `http://localhost:8082`。
-工单02~11 后端未完成，前端只定义请求结构、页面交互和占位接口。
-页面状态进入 URL hash query，保证筛选、Tab、页码可回放。
-移动端不引入 Element Plus；新增页面只使用 Vant 和项目 CSS。

---

### Task 1: API Infrastructure

**Files:**
- Create: `admin-platform/web/src/api/rideHailing/workorder01.js`
- Create: `admin-platform/web/src/api/rideHailing/placeholders.js`
- Create: `passenger-platform/web/src/utils/gatewayRequest.ts`
- Create: `passenger-platform/web/src/api/workorder01.ts`
- Create: `passenger-platform/web/src/api/placeholders.ts`
- Create: `driver-platform/web/src/utils/gatewayRequest.ts`
- Create: `driver-platform/web/src/api/workorder01.ts`
- Create: `driver-platform/web/src/api/placeholders.ts`

**Interfaces:**
- Produces admin API functions: `listCertAudits`, `getCertAudit`, `approveCertAudit`, `rejectCertAudit`, `listVehicleReviews`, `handleVehicleReview`, `callPlaceholderApi`.
- Produces mobile Gateway functions for workorder 01 and placeholder API helpers returning `{ code, data, msg }`.

- [ ] Create API files and ensure each function maps to the documented URL.
- [ ] Include `X-User-Id` support in mobile Gateway request helper.
- [ ] Add placeholder mock response functions for workorder 02~11.
- [ ] Run TypeScript/build verification after consuming pages are added.

### Task 2: Admin Workorder Console

**Files:**
- Create: `admin-platform/web/src/view/rideHailing/workorders/index.vue`
- Modify: `admin-platform/web/src/router/index.js`

**Interfaces:**
- Consumes admin APIs from Task 1.
- Produces route `/ride-hailing/workorders` with tabs for workorder 01~11.

- [ ] Build a GVA-style table page with filter form, batch buttons, table, pagination, detail drawer, and action dialog.
- [ ] Workorder 01 tab calls real admin review APIs.
- [ ] Workorder 02~11 tabs use placeholder rows and placeholder API functions.
- [ ] Store active workorder, tab, keyword, status, and page in route query.

### Task 3: Passenger Pages and Routes

**Files:**
- Create/modify passenger workorder views under `passenger-platform/web/src/views/passenger/workorderXX`.
- Modify: `passenger-platform/web/src/router/passenger.ts`
- Modify: `passenger-platform/web/src/layouts/PassengerLayout.vue`

**Interfaces:**
- Consumes passenger APIs from Task 1.
- Produces passenger routes for shuttle ticket, refund progress, coupons, performance probes, AI, and trip tracking.

- [ ] Replace new page usage with Vant-only controls.
- [ ] Add workorder 02 shuttle ticket page with calendar, seats, station ordering, duplicate-station guard, and buy dialog.
- [ ] Add workorder 03 refund progress timeline and workorder 07 coupon refund status details.
- [ ] Add workorder 08 WS/map probe page and workorder 11 tracking enhancements.
- [ ] Keep route state in hash query.

### Task 4: Driver Pages and Routes

**Files:**
- Create/modify driver workorder views under `driver-platform/web/src/views/driver/workorderXX`.
- Modify: `driver-platform/web/src/router/driver.ts`
- Modify: `driver-platform/web/src/layouts/DriverLayout.vue`

**Interfaces:**
- Consumes driver APIs from Task 1.
- Produces driver routes for today shuttle schedule, income ledger, status restriction, performance probes, AI alerts, dispatch, location, and track replay.

- [ ] Add workorder 02 today shuttle schedule page.
- [ ] Add workorder 03 income ledger page with order amount, commission, refund deduction, reward, net income.
- [ ] Add workorder 08 location/WS performance acceptance page.
- [ ] Keep existing workorder 11 dispatch/location route names working.

### Task 5: Verification

**Files:**
- No new files unless verification reveals issues.

**Commands:**
- `npm run build` in `passenger-platform/web`
- `npm run build` in `driver-platform/web`
- `npm run build` in `admin-platform/web`

- [ ] Run all three builds.
- [ ] Fix compile errors caused by this batch.
- [ ] Report any pre-existing project build blockers separately if they are outside this batch.
