package router

import (
	"ride-hailing/admin-server/router/carpool"
	"ride-hailing/admin-server/router/example"
	"ride-hailing/admin-server/router/media"
	"ride-hailing/admin-server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System   system.RouterGroup
	Example  example.RouterGroup
	Media    media.RouterGroup
	Carpool  carpool.RouterGroup
}
