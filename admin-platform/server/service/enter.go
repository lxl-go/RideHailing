package service

import (
	"ride-hailing/admin-server/service/carpool"
	"ride-hailing/admin-server/service/example"
	"ride-hailing/admin-server/service/media"
	"ride-hailing/admin-server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup   system.ServiceGroup
	ExampleServiceGroup  example.ServiceGroup
	MediaServiceGroup    media.ServiceGroup
	CarpoolServiceGroup  carpool.ServiceGroup
}
