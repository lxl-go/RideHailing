package example

import (
	api "ride-hailing/admin-server/api/v1"
)

type RouterGroup struct {
	CustomerRouter
}

var (
	exaCustomerApi = api.ApiGroupApp.ExampleApiGroup.CustomerApi
)
