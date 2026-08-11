package v1

import (
	"ride-hailing/admin-server/api/v1/carpool"
	"ride-hailing/admin-server/api/v1/example"
	"ride-hailing/admin-server/api/v1/media"
	"ride-hailing/admin-server/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup   system.ApiGroup
	ExampleApiGroup  example.ApiGroup
	MediaApiGroup    media.ApiGroup
	CarpoolApiGroup  carpool.ApiGroup
}
