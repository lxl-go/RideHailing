package auto

import (
	"context"

	"ride-hailing/admin-server/plugin/auto/initialize"
	interfaces "ride-hailing/admin-server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Dictionary(ctx)
	initialize.Gorm(ctx)
	initialize.Router(engine)
}
