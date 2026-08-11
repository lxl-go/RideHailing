package initialize

import (
	"context"
	model "ride-hailing/admin-server/model/system"
	"ride-hailing/admin-server/plugin/plugin-tool/utils"
)

func Dictionary(ctx context.Context) {
	entities := []model.SysDictionary{}
	utils.RegisterDictionaries(entities...)
}
