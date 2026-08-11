package initialize

import (
	"context"
	"fmt"

	"ride-hailing/admin-server/global"
	aiModel "ride-hailing/admin-server/plugin/ai/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(aiModel.SysCli),
		new(aiModel.SysCliApi),
		new(aiModel.SysMcp),
		new(aiModel.SysMcpApi),
	)
	if err != nil {
		err = errors.Wrap(err, "register ai plugin tables failed")
		zap.L().Error(fmt.Sprintf("%+v", err))
	}
}
