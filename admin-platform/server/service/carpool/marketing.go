package carpool

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"gorm.io/gorm"
)

const (
	couponStatusEnabled = "enabled"
	userCouponUnused    = "unused"
	userCouponUsed      = "used"
)

type MarketingService struct{}

type ReferralSummary struct {
	TotalRewards   int64 `json:"totalRewards"`
	IssuedRewards  int64 `json:"issuedRewards"`
	PendingRewards int64 `json:"pendingRewards"`
}

// WO-07 marketing: create coupon template with type-specific value validation and date scope.
func (s *MarketingService) CreateCouponTemplate(ctx context.Context, req carpoolReq.SaveCouponTemplateRequest) (*carpoolModel.CouponTemplate, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.CouponType) == "" {
		return nil, errors.New("name and couponType are required")
	}
	validFrom, err := time.ParseInLocation("2006-01-02", req.ValidFrom, time.Local)
	if err != nil {
		return nil, err
	}
	validTo, err := time.ParseInLocation("2006-01-02", req.ValidTo, time.Local)
	if err != nil {
		return nil, err
	}
	if !validTo.After(validFrom) {
		return nil, errors.New("validTo must be after validFrom")
	}
	if req.CouponType == "discount" {
		if req.DiscountRate <= 0 || req.DiscountRate >= 1 {
			return nil, errors.New("discountRate must be between 0 and 1")
		}
	} else if req.FaceValue <= 0 {
		return nil, errors.New("faceValue must be greater than 0")
	}
	template := &carpoolModel.CouponTemplate{
		CouponNo:        nextCouponNo(),
		Name:            strings.TrimSpace(req.Name),
		CouponType:      req.CouponType,
		FaceValue:       req.FaceValue,
		DiscountRate:    req.DiscountRate,
		ThresholdAmount: req.ThresholdAmount,
		ValidFrom:       validFrom,
		ValidTo:         validTo,
		CityScope:       req.CityScope,
		ServiceScope:    fallback(req.ServiceScope, "all"),
		TimeScope:       req.TimeScope,
		Stackable:       req.Stackable,
		TotalStock:      req.TotalStock,
		Status:          fallback(req.Status, couponStatusEnabled),
	}
	if err := global.GVA_DB.WithContext(ctx).Create(template).Error; err != nil {
		return nil, err
	}
	return template, nil
}

func (s *MarketingService) ListCouponTemplates(ctx context.Context, search carpoolReq.CouponTemplateSearch) ([]carpoolModel.CouponTemplate, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.CouponTemplate{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("coupon_no LIKE ? OR name LIKE ?", keyword, keyword)
	}
	if search.CouponType != "" {
		db = db.Where("coupon_type = ?", search.CouponType)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if search.CityScope != "" {
		db = db.Where("city_scope = ?", search.CityScope)
	}
	if search.ServiceType != "" {
		db = db.Where("service_scope = ? OR service_scope = ?", search.ServiceType, "all")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.CouponTemplate
	if err := db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// WO-07 marketing: issue coupon manually or from campaign/referral rules with stock validation.
func (s *MarketingService) IssueCoupon(ctx context.Context, req carpoolReq.IssueCouponRequest) (*carpoolModel.UserCouponView, error) {
	if req.CouponNo == "" {
		return nil, errors.New("couponNo and userId are required")
	}
	userID, err := parsePositiveUintString(req.UserID, "userId")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var issued carpoolModel.UserCoupon
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var template carpoolModel.CouponTemplate
		if err := tx.Where("coupon_no = ? AND status = ?", req.CouponNo, couponStatusEnabled).First(&template).Error; err != nil {
			return err
		}
		if template.TotalStock > 0 && template.IssuedCount >= template.TotalStock {
			return errors.New("coupon stock exhausted")
		}
		issued = carpoolModel.UserCoupon{
			CouponCode: nextCouponCode(),
			CouponNo:   template.CouponNo,
			UserID:     userID,
			UserType:   fallback(req.UserType, "passenger"),
			Source:     fallback(req.Source, "manual"),
			Status:     userCouponUnused,
			Operator:   req.Operator,
			IssuedAt:   now,
		}
		if err := tx.Create(&issued).Error; err != nil {
			return err
		}
		return tx.Model(&carpoolModel.CouponTemplate{}).Where("coupon_no = ?", template.CouponNo).
			Update("issued_count", gorm.Expr("issued_count + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}
	return userCouponViewFromModel(issued), nil
}

// WO-07 marketing: redeem one coupon per order by default and calculate discount from template strategy.
func (s *MarketingService) RedeemCoupon(ctx context.Context, req carpoolReq.RedeemCouponRequest) (*carpoolModel.UserCoupon, error) {
	if req.CouponCode == "" || req.OrderNo == "" || req.OrderAmount <= 0 {
		return nil, errors.New("couponCode, orderNo and orderAmount are required")
	}
	var redeemed carpoolModel.UserCoupon
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCoupon carpoolModel.UserCoupon
		if err := tx.Where("coupon_code = ? AND status = ?", req.CouponCode, userCouponUnused).First(&userCoupon).Error; err != nil {
			return err
		}
		var template carpoolModel.CouponTemplate
		if err := tx.Where("coupon_no = ?", userCoupon.CouponNo).First(&template).Error; err != nil {
			return err
		}
		if time.Now().Before(template.ValidFrom) || time.Now().After(template.ValidTo) {
			return errors.New("coupon is not valid now")
		}
		if req.OrderAmount < template.ThresholdAmount {
			return errors.New("order amount does not meet coupon threshold")
		}
		discount := calculateCouponDiscount(template, req.OrderAmount)
		if err := tx.Model(&userCoupon).Updates(map[string]interface{}{
			"status":          userCouponUsed,
			"order_no":        req.OrderNo,
			"order_amount":    req.OrderAmount,
			"discount_amount": discount,
			"used_at":         time.Now(),
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&carpoolModel.CouponTemplate{}).Where("coupon_no = ?", template.CouponNo).
			Update("used_count", gorm.Expr("used_count + ?", 1)).Error; err != nil {
			return err
		}
		redeemed = userCoupon
		redeemed.Status = userCouponUsed
		redeemed.OrderNo = req.OrderNo
		redeemed.OrderAmount = req.OrderAmount
		redeemed.DiscountAmount = discount
		redeemed.UsedAt = time.Now()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &redeemed, nil
}

func (s *MarketingService) DeleteCouponTemplate(ctx context.Context, couponNo string) error {
	if couponNo == "" {
		return errors.New("couponNo is required")
	}
	var issued int64
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.UserCoupon{}).Where("coupon_no = ?", couponNo).Count(&issued).Error; err != nil {
		return err
	}
	if issued > 0 {
		return errors.New("coupon template already issued")
	}
	return global.GVA_DB.WithContext(ctx).Where("coupon_no = ?", couponNo).Delete(&carpoolModel.CouponTemplate{}).Error
}

func (s *MarketingService) ListUserCoupons(ctx context.Context, search carpoolReq.UserCouponSearch) ([]carpoolModel.UserCouponView, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.UserCoupon{})
	if search.CouponNo != "" {
		db = db.Where("coupon_no = ?", search.CouponNo)
	}
	if search.CouponCode != "" {
		db = db.Where("coupon_code LIKE ?", "%"+strings.TrimSpace(search.CouponCode)+"%")
	}
	if strings.TrimSpace(search.UserID) != "" {
		userID, err := parsePositiveUintString(search.UserID, "userId")
		if err != nil {
			return nil, 0, err
		}
		db = db.Where("user_id = ?", userID)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if search.Source != "" {
		db = db.Where("source = ?", search.Source)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.UserCoupon
	if err := db.Order("issued_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return userCouponViewsFromModels(list), total, nil
}

func (s *MarketingService) ListCampaigns(ctx context.Context, search carpoolReq.MarketingCampaignSearch) ([]carpoolModel.MarketingCampaign, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.MarketingCampaign{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("campaign_no LIKE ? OR name LIKE ?", keyword, keyword)
	}
	if search.Channel != "" {
		db = db.Where("channel = ?", search.Channel)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.MarketingCampaign
	if err := db.Order("start_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *MarketingService) GetReferralSummary(ctx context.Context) (*ReferralSummary, error) {
	summary := &ReferralSummary{}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.ReferralReward{}).Count(&summary.TotalRewards).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.ReferralReward{}).Where("reward_status = ?", "issued").Count(&summary.IssuedRewards).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.ReferralReward{}).Where("reward_status = ?", "pending").Count(&summary.PendingRewards).Error; err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *MarketingService) ExportTaskID(ctx context.Context) string {
	return fmt.Sprintf("MKT-EXP-%d", time.Now().UnixNano())
}

func SeedMarketingDefaults(db *gorm.DB) error {
	var count int64
	if err := db.Model(&carpoolModel.CouponTemplate{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	templates := []carpoolModel.CouponTemplate{
		{ID: 71001, CouponNo: "CP202607290001", Name: "New Passenger Cash", CouponType: "cash", FaceValue: 20, ThresholdAmount: 50, ValidFrom: now.AddDate(0, 0, -3), ValidTo: now.AddDate(0, 1, 0), CityScope: "Beijing", ServiceScope: "carpool", TotalStock: 1000, Status: couponStatusEnabled},
		{ID: 71002, CouponNo: "CP202607290002", Name: "Shuttle 20 Percent Off", CouponType: "discount", DiscountRate: 0.8, ThresholdAmount: 80, ValidFrom: now.AddDate(0, 0, -3), ValidTo: now.AddDate(0, 1, 0), CityScope: "Beijing", ServiceScope: "shuttle", TotalStock: 500, Status: couponStatusEnabled},
		{ID: 71003, CouponNo: "CP202607290003", Name: "Free Short Ride", CouponType: "free_ride", FaceValue: 30, ThresholdAmount: 0, ValidFrom: now.AddDate(0, 0, -1), ValidTo: now.AddDate(0, 0, 14), CityScope: "Beijing", ServiceScope: "all", TotalStock: 200, Status: couponStatusEnabled},
	}
	if err := db.Create(&templates).Error; err != nil {
		return err
	}
	campaigns := []carpoolModel.MarketingCampaign{
		{ID: 72001, CampaignNo: "MK202607290001", Name: "Summer Referral Campaign", Channel: "social", CouponNo: "CP202607290001", StartAt: now.AddDate(0, 0, -1), EndAt: now.AddDate(0, 0, 14), Status: "running"},
		{ID: 72002, CampaignNo: "MK202607290002", Name: "Station Banner Campaign", Channel: "banner", CouponNo: "CP202607290002", StartAt: now.AddDate(0, 0, -2), EndAt: now.AddDate(0, 0, 7), Status: "running"},
	}
	if err := db.Create(&campaigns).Error; err != nil {
		return err
	}
	return db.Create(&[]carpoolModel.ReferralReward{
		{ID: 73001, ReferrerID: 1001, InviteeID: 1011, CouponNo: "CP202607290001", RewardStatus: "issued"},
		{ID: 73002, ReferrerID: 1002, InviteeID: 1012, CouponNo: "CP202607290001", RewardStatus: "pending"},
	}).Error
}

func calculateCouponDiscount(template carpoolModel.CouponTemplate, orderAmount float64) float64 {
	switch template.CouponType {
	case "cash", "free_ride":
		return roundMarketingAmount(minFloat(template.FaceValue, orderAmount))
	case "discount":
		if template.DiscountRate <= 0 || template.DiscountRate >= 1 {
			return 0
		}
		return roundMarketingAmount(orderAmount * (1 - template.DiscountRate))
	default:
		return 0
	}
}

func nextCouponCode() string {
	return fmt.Sprintf("UC%d", time.Now().UnixNano())
}

func nextCouponNo() string {
	return fmt.Sprintf("CP%d", time.Now().UnixNano())
}

func roundMarketingAmount(v float64) float64 {
	return math.Round(v*100) / 100
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func userCouponViewsFromModels(list []carpoolModel.UserCoupon) []carpoolModel.UserCouponView {
	views := make([]carpoolModel.UserCouponView, 0, len(list))
	for _, item := range list {
		views = append(views, *userCouponViewFromModel(item))
	}
	return views
}

func userCouponViewFromModel(item carpoolModel.UserCoupon) *carpoolModel.UserCouponView {
	return &carpoolModel.UserCouponView{
		ID:             strconv.FormatUint(item.ID, 10),
		CouponCode:     item.CouponCode,
		CouponNo:       item.CouponNo,
		UserID:         strconv.FormatUint(item.UserID, 10),
		UserType:       item.UserType,
		Source:         item.Source,
		Status:         item.Status,
		OrderNo:        item.OrderNo,
		OrderAmount:    item.OrderAmount,
		DiscountAmount: item.DiscountAmount,
		Operator:       item.Operator,
		IssuedAt:       item.IssuedAt,
		UsedAt:         item.UsedAt,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}
