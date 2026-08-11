# Workorder05 Person Management Design

## Goal

Build workorder05 as a real management-side person module covering staff, drivers, passengers, multi-role assignment, import validation, duplicate prevention, sensitive data masking, batch enable/disable/delete, and an admin UI.

## Scope

- Admin platform is the primary delivery surface.
- Passenger and driver platforms are not changed in this workorder except for compatibility verification.
- GVA `sys_user` login, authority, department, and position modules remain untouched. Workorder05 owns ride-hailing business people, not platform administrator accounts.
- Dispatcher independent login and dispatch-specific permissions stay out of scope until workorder09/workorder11.

## Architecture

Add a new `person` module under the existing admin carpool package. It follows the current GVA pattern:

- `model/carpool/person.go`
- `model/carpool/request/person.go`
- `service/carpool/person.go`
- `api/v1/carpool/person.go`
- `router/carpool/person.go`
- `web/src/api/rideHailing/workorder05.js`
- `web/src/view/admin/workorder05/person.vue`

The module stores one person profile and many role rows. Role assignment is handled as a replace-all operation inside a transaction so one person can own multiple roles without partial updates.

## Data Model

### `person_profile`

- `id`: uint64 primary key.
- `person_no`: unique business number.
- `person_type`: `staff`, `driver`, or `passenger`.
- `name`, `phone`, `phone_hash`, `email`.
- `id_card_no`, `id_card_hash`.
- `driver_license_no`, `vehicle_no`, `vehicle_type`.
- `common_address`, `payment_preference`, `rating`.
- `status`: `enabled`, `disabled`, `deleted`.
- `register_date`, `disabled_reason`, `created_at`, `updated_at`.

### `person_role`

- `id`: uint64 primary key.
- `person_id`, `role_code`, `role_name`.
- Role codes: `staff`, `passenger`, `shuttle_driver`, `pickup_driver`, `carpool_driver`, `dispatcher`, `ticket_checker`, `contact`.

### `person_import_batch`

- `id`, `batch_no`, `source_type`, `total`, `success_count`, `error_count`, `status`, `operator`, `created_at`.

### `person_import_error`

- `id`, `batch_no`, `row_no`, `field`, `message`, `raw_data`, `created_at`.

## Business Rules

- One person can own multiple roles.
- Phone and ID card are validated at service/API boundary.
- Duplicate checks use normalized phone/id-card hashes.
- API response returns masked phone and masked ID card only.
- Import is transactional: if any row fails validation or duplicate check, no person rows are inserted, a failed batch and row errors are persisted.
- Batch operation supports enable, disable, and soft delete. Deleted records are excluded by default list filters.

## API Design

- `GET /carpool/person/list`
- `GET /carpool/person/:id`
- `POST /carpool/person`
- `PUT /carpool/person/:id`
- `POST /carpool/person/roles`
- `POST /carpool/person/batch/status`
- `POST /carpool/person/import/preview`
- `POST /carpool/person/import/commit`
- `GET /carpool/person/import/errors`
- `POST /carpool/person/export`

## Frontend

Add a single management page with three tabs: staff, driver, passenger. The page provides:

- Search by keyword, type, status, and role.
- Table with masked phone and ID card.
- Create/edit dialog.
- Detail drawer with roles and vehicle/license fields.
- Role assignment dialog with multi-select.
- Batch enable/disable/delete buttons.
- Import preview/commit dialog using pasted JSON rows for current backend contract.
- Import error table.
- Export button.

## Testing

- Service test verifies:
  - Create person masks sensitive data.
  - Duplicate phone/id-card is rejected.
  - Multi-role assignment replaces roles.
  - Batch status updates multiple people.
  - Import preview detects invalid phone/id-card and duplicates.
  - Import commit is transactional and creates no people when any row is invalid.

## Notes

The workspace root is not recognized as a git repository, so design commit is skipped.
