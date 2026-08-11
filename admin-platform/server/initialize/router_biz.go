package initialize

import (
	"github.com/gin-gonic/gin"
	"ride-hailing/admin-server/router"
)

// 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	holder(publicGroup, privateGroup)

	carpoolRouter := router.RouterGroupApp.Carpool
	carpoolRouter.InitReviewRouter(privateGroup)
	carpoolRouter.InitTripRouter(privateGroup)
	carpoolRouter.InitShuttleRouter(privateGroup)
	carpoolRouter.InitFinanceRouter(privateGroup)
	carpoolRouter.InitOrderRouter(privateGroup)
	carpoolRouter.InitPersonRouter(privateGroup)
	carpoolRouter.InitAnalyticsRouter(privateGroup)
	carpoolRouter.InitMarketingRouter(privateGroup)
	carpoolRouter.InitPerformanceRouter(privateGroup)
	carpoolRouter.InitAIRouter(privateGroup, publicGroup)
	carpoolRouter.InitDispatchRouter(privateGroup)
}
