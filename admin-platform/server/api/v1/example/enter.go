package example

import "ride-hailing/admin-server/service"

type ApiGroup struct {
	CustomerApi
}

var (
	customerService = service.ServiceGroupApp.ExampleServiceGroup.CustomerService
)
