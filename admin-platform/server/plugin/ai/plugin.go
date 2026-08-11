package ai

import (
	"context"

	"ride-hailing/admin-server/plugin/ai/initialize"
	interfaces "ride-hailing/admin-server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(group *gin.Engine) {
	ctx := context.Background()
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Gorm(ctx)
	initialize.Router(group)
}
