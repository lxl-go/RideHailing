package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type MarketingRouter struct{}

func (r *MarketingRouter) InitMarketingRouter(Router *gin.RouterGroup) {
	marketingRouter := Router.Group("carpool/marketing").Use(middleware.OperationRecord())
	marketingRouterWithoutRecord := Router.Group("carpool/marketing")
	{
		marketingRouter.POST("coupon/template", marketingApi.CreateCouponTemplate)
		marketingRouter.POST("coupon/issue", marketingApi.IssueCoupon)
		marketingRouter.POST("coupon/redeem", marketingApi.RedeemCoupon)
		marketingRouter.DELETE("coupon/template/:couponNo", marketingApi.DeleteCouponTemplate)
		marketingRouter.POST("export", marketingApi.ExportMarketing)
	}
	{
		marketingRouterWithoutRecord.GET("coupon/template/list", marketingApi.ListCouponTemplates)
		marketingRouterWithoutRecord.GET("coupon/user/list", marketingApi.ListUserCoupons)
		marketingRouterWithoutRecord.GET("campaign/list", marketingApi.ListCampaigns)
		marketingRouterWithoutRecord.GET("referral/summary", marketingApi.GetReferralSummary)
	}
}
