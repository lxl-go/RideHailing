package initialize

import (
	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/middleware"
	"ride-hailing/admin-server/plugin/auto/router"

	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	InitializeRouter(engine)
}

func InitializeRouter(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())

	router.RouterGroupApp.InitAutoCodeRouter(private, public)
	router.RouterGroupApp.InitAutoCodeHistoryRouter(private)
}
