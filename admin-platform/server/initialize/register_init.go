package initialize

import (
	_ "ride-hailing/admin-server/source/media"
	_ "ride-hailing/admin-server/source/system"
)

func init() {
	// do nothing,only import source package so that inits can be registered
}
