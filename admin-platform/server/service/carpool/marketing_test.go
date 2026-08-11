package carpool

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

func TestMarketingServiceCouponLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.CouponTemplate{},
		&carpoolModel.UserCoupon{},
		&carpoolModel.MarketingCampaign{},
		&carpoolModel.ReferralReward{},
	))
	global.GVA_DB = db

	now := time.Now()
	start := now.AddDate(0, 0, -1)
	end := now.AddDate(0, 0, 7)
	require.NoError(t, db.Create(&[]carpoolModel.CouponTemplate{
		{
			ID: 1001, CouponNo: "CP202607290001", Name: "First Ride Cash", CouponType: "cash",
			FaceValue: 20, ThresholdAmount: 50, ValidFrom: start, ValidTo: end,
			CityScope: "Beijing", ServiceScope: "carpool", TotalStock: 100, IssuedCount: 0, Status: "enabled",
		},
		{
			ID: 1002, CouponNo: "CP202607290002", Name: "Airport Discount", CouponType: "discount",
			DiscountRate: 0.8, ThresholdAmount: 80, ValidFrom: start, ValidTo: end,
			CityScope: "Beijing", ServiceScope: "shuttle", TotalStock: 20, IssuedCount: 0, Status: "enabled",
		},
	}).Error)

	service := MarketingService{}
	ctx := context.Background()

	created, err := service.CreateCouponTemplate(ctx, carpoolReq.SaveCouponTemplateRequest{
		Name: "Night Coupon", CouponType: "cash", FaceValue: 15, ThresholdAmount: 40,
		ValidFrom: "2026-07-01", ValidTo: "2026-08-01", CityScope: "Beijing", ServiceScope: "carpool", TotalStock: 30, Status: "enabled",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.CouponNo)
	require.Equal(t, "Night Coupon", created.Name)

	templates, total, err := service.ListCouponTemplates(ctx, carpoolReq.CouponTemplateSearch{Keyword: "Cash", CouponType: "cash"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "First Ride Cash", templates[0].Name)

	issued, err := service.IssueCoupon(ctx, carpoolReq.IssueCouponRequest{
		CouponNo: "CP202607290001", UserID: "9001", UserType: "passenger", Source: "manual", Operator: "admin",
	})
	require.NoError(t, err)
	require.Equal(t, "unused", issued.Status)
	require.Equal(t, "9001", issued.UserID)

	var template carpoolModel.CouponTemplate
	require.NoError(t, db.Where("coupon_no = ?", "CP202607290001").First(&template).Error)
	require.EqualValues(t, 1, template.IssuedCount)

	redeemed, err := service.RedeemCoupon(ctx, carpoolReq.RedeemCouponRequest{
		CouponCode: issued.CouponCode, OrderNo: "OD202607290001", OrderAmount: 88,
	})
	require.NoError(t, err)
	require.Equal(t, "used", redeemed.Status)
	require.Equal(t, "OD202607290001", redeemed.OrderNo)
	require.Equal(t, 20.0, redeemed.DiscountAmount)

	userCoupons, total, err := service.ListUserCoupons(ctx, carpoolReq.UserCouponSearch{UserID: "9001", Status: "used"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, issued.CouponCode, userCoupons[0].CouponCode)

	err = service.DeleteCouponTemplate(ctx, "CP202607290001")
	require.Error(t, err)
	require.Contains(t, err.Error(), "coupon template already issued")
}

func TestMarketingServiceIssueAndQueryCouponWithStringUserID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.CouponTemplate{},
		&carpoolModel.UserCoupon{},
	))
	global.GVA_DB = db

	now := time.Date(2026, 8, 4, 0, 0, 0, 0, time.Local)
	require.NoError(t, db.Create(&carpoolModel.CouponTemplate{
		CouponNo: "CP-LARGE-ID", Name: "Large ID Coupon", CouponType: "cash",
		FaceValue: 10, ValidFrom: now.AddDate(0, 0, -1), ValidTo: now.AddDate(0, 0, 7),
		ServiceScope: "all", TotalStock: 10, Status: "enabled",
	}).Error)

	service := MarketingService{}
	userID := "9007199254740993"
	issued, err := service.IssueCoupon(context.Background(), carpoolReq.IssueCouponRequest{
		CouponNo: "CP-LARGE-ID", UserID: userID, UserType: "passenger", Source: "manual", Operator: "admin",
	})
	require.NoError(t, err)
	require.Equal(t, userID, issued.UserID)

	list, total, err := service.ListUserCoupons(context.Background(), carpoolReq.UserCouponSearch{UserID: userID})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, userID, list[0].UserID)
}

func TestMarketingServiceCampaignAndReferralSummary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.CouponTemplate{},
		&carpoolModel.UserCoupon{},
		&carpoolModel.MarketingCampaign{},
		&carpoolModel.ReferralReward{},
	))
	global.GVA_DB = db

	now := time.Date(2026, 7, 29, 10, 0, 0, 0, time.Local)
	require.NoError(t, db.Create(&[]carpoolModel.MarketingCampaign{
		{ID: 2001, CampaignNo: "MK202607290001", Name: "Summer Promo", Channel: "social", CouponNo: "CP202607290001", StartAt: now.AddDate(0, 0, -1), EndAt: now.AddDate(0, 0, 7), Status: "running"},
		{ID: 2002, CampaignNo: "MK202607290002", Name: "Station Banner", Channel: "banner", CouponNo: "CP202607290002", StartAt: now.AddDate(0, 0, -2), EndAt: now.AddDate(0, 0, 1), Status: "paused"},
	}).Error)
	require.NoError(t, db.Create(&[]carpoolModel.ReferralReward{
		{ID: 3001, ReferrerID: 9001, InviteeID: 9101, CouponNo: "CP202607290001", RewardStatus: "issued", CreatedAt: now},
		{ID: 3002, ReferrerID: 9002, InviteeID: 9102, CouponNo: "CP202607290002", RewardStatus: "pending", CreatedAt: now},
	}).Error)

	service := MarketingService{}
	ctx := context.Background()

	campaigns, total, err := service.ListCampaigns(ctx, carpoolReq.MarketingCampaignSearch{Status: "running"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "Summer Promo", campaigns[0].Name)

	summary, err := service.GetReferralSummary(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 2, summary.TotalRewards)
	require.EqualValues(t, 1, summary.IssuedRewards)
	require.EqualValues(t, 1, summary.PendingRewards)
}
