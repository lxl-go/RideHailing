package initialize

import (
	"os"

	adapter "github.com/casbin/gorm-adapter/v3"
	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/model/carpool"
	"ride-hailing/admin-server/model/example"
	"ride-hailing/admin-server/model/media"
	"ride-hailing/admin-server/model/system"
	carpoolService "ride-hailing/admin-server/service/carpool"
	"ride-hailing/admin-server/utils/logger"

	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	switch global.GVA_CONFIG.System.DbType {
	case "mysql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mysql.Dbname
		return GormMysql()
	case "pgsql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Pgsql.Dbname
		return GormPgSql()
	case "oracle":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Oracle.Dbname
		return GormOracle()
	case "mssql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mssql.Dbname
		return GormMssql()
	case "sqlite":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Sqlite.Dbname
		return GormSqlite()
	default:
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mysql.Dbname
		return GormMysql()
	}
}

func RegisterTables() {
	if global.GVA_CONFIG.System.DisableAutoMigrate {
		logger.Bg().Mod("system").Info("auto-migrate is disabled, skipping table registration")
		return
	}

	db := global.GVA_DB
	err := db.AutoMigrate(

		system.SysApi{},
		system.SysIgnoreApi{},
		system.SysUser{},
		system.SysBaseMenu{},
		system.JwtBlacklist{},
		system.SysAuthority{},
		system.SysDepartment{},
		system.SysPosition{},
		system.SysDataAccessLog{},
		system.SysAuthorityDepartment{},
		system.SysDictionary{},
		system.SysOperationRecord{},
		system.SysAutoCodeHistory{},
		system.SysDictionaryDetail{},
		system.SysBaseMenuParameter{},
		system.SysBaseMenuBtn{},
		system.SysAuthorityBtn{},
		system.SysAutoCodePackage{},
		system.SysExportTemplate{},
		system.Condition{},
		system.JoinTemplate{},
		system.SysParams{},
		system.SysSecurityConfig{},
		system.SysVersion{},
		system.SysError{},
		system.SysApiToken{},
		system.SysLoginLog{},
		system.SysTimedTask{},
		system.SysTimedTaskLog{},

		carpool.CertificationAudit{},
		carpool.DriverProfile{},
		carpool.DriverCertification{},
		carpool.DriverVehicle{},
		carpool.VehicleInfo{},
		carpool.Trip{},
		carpool.ShuttleLine{},
		carpool.FinanceTransaction{},
		carpool.FinanceRefund{},
		carpool.OrderMain{},
		carpool.OrderRefund{},
		carpool.OrderStatusHistory{},
		carpool.PersonProfile{},
		carpool.PersonRole{},
		carpool.PersonImportBatch{},
		carpool.PersonImportError{},
		carpool.CouponTemplate{},
		carpool.UserCoupon{},
		carpool.MarketingCampaign{},
		carpool.ReferralReward{},
		carpool.PerformanceReport{},
		carpool.PerformanceScenario{},
		carpool.AiConversationLog{},
		carpool.AiRoutePlanLog{},
		carpool.AiFloodReport{},
		carpool.AiFallbackTemplate{},
		carpool.OrderDispatchAudit{},
		carpool.DispatchConfig{},
		carpool.DispatchConfigVersion{},
		carpool.DriverLocationPoint{},
		carpool.RealtimeMessage{},

		example.ExaCustomer{},
		media.MediaUpload{},
		media.MediaUploadChunk{},
		media.FileUploadAndDownload{},
		media.AttachmentCategory{},
	)
	if err != nil {
		logger.Bg().Mod("system").Err(err).Error("register table failed")
		os.Exit(1)
	}
	if err = carpoolServiceSeed(db); err != nil {
		logger.Bg().Mod("system").Err(err).Error("seed carpool biz data failed")
		os.Exit(1)
	}
	if err = ensureRideHailingAdminPermissions(db); err != nil {
		logger.Bg().Mod("system").Err(err).Error("seed ride hailing admin permissions failed")
		os.Exit(1)
	}

	err = bizModel()

	if err != nil {
		logger.Bg().Mod("system").Err(err).Error("register biz_table failed")
		os.Exit(1)
	}
	logger.Bg().Mod("system").Info("register table success")
}

func carpoolServiceSeed(db *gorm.DB) error {
	if err := carpoolService.SeedOrderDefaults(db); err != nil {
		return err
	}
	if err := carpoolService.SeedPersonDefaults(db); err != nil {
		return err
	}
	if err := carpoolService.SeedMarketingDefaults(db); err != nil {
		return err
	}
	return carpoolService.SeedPerformanceDefaults(db)
}

func ensureRideHailingAdminPermissions(db *gorm.DB) error {
	apis := []system.SysApi{
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/analytics/dashboard", Description: "运营概览"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/analytics/order-volume", Description: "订单量趋势"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/analytics/classification", Description: "订单分类"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/analytics/conversion", Description: "转化分析"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/analytics/repurchase", Description: "复购分析"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/analytics/export", Description: "导出分析"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/order/overview", Description: "订单概览"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/order/list", Description: "订单列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/order/refund/list", Description: "退款列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/order/:orderNo", Description: "订单详情"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/order/:orderNo/history", Description: "订单状态历史"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/order/refund/apply", Description: "申请退款"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/order/refund/review", Description: "审核退款"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/order/refund/batch", Description: "批量退款"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/order/export", Description: "导出订单"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/person/list", Description: "人员列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/person/:id", Description: "人员详情"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/driver/stats", Description: "司机统计"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/passenger/stats", Description: "乘客统计"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person", Description: "创建人员"},
		{ApiGroup: "网约车", Method: "PUT", Path: "/carpool/person/:id", Description: "更新人员"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person/roles", Description: "分配人员角色"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person/batch/status", Description: "批量更新人员状态"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person/driver/batch/delete", Description: "批量删除司机"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person/import/preview", Description: "人员导入预览"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person/import/commit", Description: "人员导入提交"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/person/import/errors", Description: "人员导入错误"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/person/export", Description: "导出人员"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/review/list", Description: "认证审核列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/review/:id", Description: "认证审核详情"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/review/:id/approve", Description: "认证审核通过"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/review/:id/reject", Description: "认证审核拒绝"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/review/vehicle/list", Description: "车辆审核列表"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/review/vehicle/:id/action", Description: "车辆审核处理"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/trip/list", Description: "行程列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/trip/:id", Description: "行程详情"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/trip/:id/deactivate", Description: "停用行程"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/shuttle/line/list", Description: "班车线路列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/shuttle/line/:id", Description: "班车线路详情"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/shuttle/line", Description: "创建班车线路"},
		{ApiGroup: "网约车", Method: "PUT", Path: "/carpool/shuttle/line/:id", Description: "更新班车线路"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/shuttle/line/publish", Description: "发布班车线路"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/shuttle/line/export", Description: "导出班车线路"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/finance/summary", Description: "财务汇总"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/finance/transaction/list", Description: "交易列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/finance/refund/list", Description: "财务退款列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/finance/abnormal/list", Description: "异常交易列表"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/finance/export", Description: "导出财务"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/marketing/coupon/template/list", Description: "优惠券模板列表"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/marketing/coupon/template", Description: "创建优惠券模板"},
		{ApiGroup: "网约车", Method: "DELETE", Path: "/carpool/marketing/coupon/template/:couponNo", Description: "删除优惠券模板"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/marketing/coupon/issue", Description: "发券"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/marketing/coupon/redeem", Description: "核销优惠券"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/marketing/coupon/user/list", Description: "用户优惠券列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/marketing/campaign/list", Description: "活动列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/marketing/referral/summary", Description: "邀请汇总"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/marketing/export", Description: "导出营销"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/performance/summary", Description: "性能汇总"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/performance/report/list", Description: "性能报告列表"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/performance/report", Description: "创建性能报告"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/performance/scenario/list", Description: "性能场景列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/performance/runtime", Description: "运行快照"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/performance/export", Description: "导出性能报告"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/ai/summary", Description: "AI 汇总"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/ai/chat", Description: "AI 对话"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/ai/rain-route", Description: "雨天路线规划"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/ai/chat-with-route", Description: "带路线 AI 对话"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/ai/flood-report", Description: "积水上报"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/ai/flood-report/audit", Description: "积水审核"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/ai/conversation/list", Description: "AI 会话列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/ai/route-plan/list", Description: "路线规划列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/ai/flood-report/list", Description: "积水上报列表"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/ai/export", Description: "导出 AI 数据"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/dispatch/order/list", Description: "派单列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/dispatch/order/:id", Description: "派单详情"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/order/:id/cancel", Description: "取消派单"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/order/:id/reassign", Description: "改派"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/score", Description: "司机评分"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/dispatch/config/list", Description: "派单配置列表"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/config", Description: "保存派单配置"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/config/:id/publish", Description: "发布派单配置"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/config/:id/rollback", Description: "回滚派单配置"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/dispatch/audit/list", Description: "派单审计列表"},
		{ApiGroup: "网约车", Method: "GET", Path: "/carpool/dispatch/track/replay", Description: "轨迹回放"},
		{ApiGroup: "网约车", Method: "POST", Path: "/carpool/dispatch/export", Description: "导出派单"},
		{ApiGroup: "平台治理", Method: "GET", Path: "/system/gva-governance/summary", Description: "治理汇总"},
		{ApiGroup: "平台治理", Method: "GET", Path: "/system/gva-governance/routes", Description: "路由快照"},
		{ApiGroup: "平台治理", Method: "GET", Path: "/system/gva-governance/audit", Description: "审计快照"},
		{ApiGroup: "平台治理", Method: "GET", Path: "/system/gva-governance/datasource", Description: "数据源快照"},
		{ApiGroup: "平台治理", Method: "GET", Path: "/system/gva-governance/timed-task", Description: "定时任务快照"},
		{ApiGroup: "平台治理", Method: "POST", Path: "/system/gva-governance/export", Description: "导出治理数据"},
	}

	for _, api := range apis {
		if err := db.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
			return err
		}
		for _, authorityID := range []string{"888", "8881", "9528"} {
			rule := adapter.CasbinRule{Ptype: "p", V0: authorityID, V1: api.Path, V2: api.Method}
			if err := db.Where(rule).FirstOrCreate(&rule).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
