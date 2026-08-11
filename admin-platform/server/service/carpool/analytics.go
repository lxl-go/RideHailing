package carpool

import (
	"context"
	"fmt"
	"math"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"gorm.io/gorm"
)

type AnalyticsService struct{}

type AnalyticsDashboard struct {
	TodayOrderCount  int64   `json:"todayOrderCount"`
	MonthOrderCount  int64   `json:"monthOrderCount"`
	MonthRevenue     float64 `json:"monthRevenue"`
	ActiveDrivers    int64   `json:"activeDrivers"`
	ActivePassengers int64   `json:"activePassengers"`
	ConversionRate   float64 `json:"conversionRate"`
}

type OrderVolumeResult struct {
	Categories  []string  `json:"categories"`
	TotalOrders []int64   `json:"totalOrders"`
	ValidOrders []int64   `json:"validOrders"`
	GrowthRates []float64 `json:"growthRates"`
}

type OrderClassificationResult struct {
	ValidOrders   int64 `json:"validOrders"`
	InvalidOrders int64 `json:"invalidOrders"`
	CouponOrders  int64 `json:"couponOrders"`
}

type ConversionResult struct {
	RegisteredUsers int64   `json:"registeredUsers"`
	PurchasedUsers  int64   `json:"purchasedUsers"`
	ConversionRate  float64 `json:"conversionRate"`
}

type RepurchaseResult struct {
	FirstOrderUsers   int64              `json:"firstOrderUsers"`
	RepurchaseUsers   int64              `json:"repurchaseUsers"`
	RepurchaseRate    float64            `json:"repurchaseRate"`
	AvgRepurchaseDays float64            `json:"avgRepurchaseDays"`
	RepurchaseBuckets []RepurchaseBucket `json:"repurchaseBuckets"`
}

type RepurchaseBucket struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// WO-06 数据分析：管理端只读 KPI 聚合，不修改源业务数据。
func (s *AnalyticsService) GetDashboard(ctx context.Context, search carpoolReq.AnalyticsSearch) (*AnalyticsDashboard, error) {
	now := time.Now()
	start, end, err := analyticsRange(search, time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local), now)
	if err != nil {
		return nil, err
	}
	todayStart := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)

	var result AnalyticsDashboard
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", todayStart, todayEnd, validOrderStatuses()).Count(&result.TodayOrderCount).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Count(&result.MonthOrderCount).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Select("COALESCE(SUM(pay_amount),0)").Scan(&result.MonthRevenue).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Distinct("driver_id").Count(&result.ActiveDrivers).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Distinct("passenger_id").Count(&result.ActivePassengers).Error; err != nil {
		return nil, err
	}
	conversion, err := s.GetConversion(ctx, search)
	if err != nil {
		return nil, err
	}
	result.MonthRevenue = roundAnalytics(result.MonthRevenue)
	result.ConversionRate = conversion.ConversionRate
	return &result, nil
}

// WO-06 数据分析：订单量日趋势，后续可扩展周/月/季/年分桶。
func (s *AnalyticsService) GetOrderVolume(ctx context.Context, search carpoolReq.AnalyticsSearch) (*OrderVolumeResult, error) {
	start, end, err := analyticsRange(search, time.Now().AddDate(0, 0, -6), time.Now())
	if err != nil {
		return nil, err
	}
	result := &OrderVolumeResult{}
	var previous int64
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		next := day.AddDate(0, 0, 1)
		var total int64
		var valid int64
		if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ?", day, next).Count(&total).Error; err != nil {
			return nil, err
		}
		if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", day, next, validOrderStatuses()).Count(&valid).Error; err != nil {
			return nil, err
		}
		result.Categories = append(result.Categories, day.Format("2006-01-02"))
		result.TotalOrders = append(result.TotalOrders, total)
		result.ValidOrders = append(result.ValidOrders, valid)
		result.GrowthRates = append(result.GrowthRates, growthRate(previous, total))
		previous = total
	}
	return result, nil
}

func (s *AnalyticsService) GetOrderClassification(ctx context.Context, search carpoolReq.AnalyticsSearch) (*OrderClassificationResult, error) {
	start, end, err := analyticsRange(search, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		return nil, err
	}
	result := &OrderClassificationResult{}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Count(&result.ValidOrders).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, invalidOrderStatuses()).Count(&result.InvalidOrders).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AnalyticsService) GetConversion(ctx context.Context, search carpoolReq.AnalyticsSearch) (*ConversionResult, error) {
	start, end, err := analyticsRange(search, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		return nil, err
	}
	result := &ConversionResult{}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PersonProfile{}).Where("person_type = ?", "passenger").Count(&result.RegisteredUsers).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Distinct("passenger_id").Count(&result.PurchasedUsers).Error; err != nil {
		return nil, err
	}
	result.ConversionRate = percent(result.PurchasedUsers, result.RegisteredUsers)
	return result, nil
}

func (s *AnalyticsService) GetRepurchase(ctx context.Context, search carpoolReq.AnalyticsSearch) (*RepurchaseResult, error) {
	start, end, err := analyticsRange(search, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		return nil, err
	}
	var orders []carpoolModel.OrderMain
	if err := orderQuery(ctx, search).Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, validOrderStatuses()).Order("passenger_id ASC, created_at ASC").Find(&orders).Error; err != nil {
		return nil, err
	}
	byPassenger := map[uint64][]carpoolModel.OrderMain{}
	for _, order := range orders {
		byPassenger[order.PassengerID] = append(byPassenger[order.PassengerID], order)
	}

	result := &RepurchaseResult{FirstOrderUsers: int64(len(byPassenger))}
	var totalDays float64
	for _, items := range byPassenger {
		if len(items) < 2 {
			continue
		}
		result.RepurchaseUsers++
		days := items[1].CreatedAt.Sub(items[0].CreatedAt).Hours() / 24
		totalDays += days
	}
	result.RepurchaseRate = percent(result.RepurchaseUsers, result.FirstOrderUsers)
	if result.RepurchaseUsers > 0 {
		result.AvgRepurchaseDays = roundAnalytics(totalDays / float64(result.RepurchaseUsers))
	}
	result.RepurchaseBuckets = []RepurchaseBucket{
		{Name: "7天内", Count: 0},
		{Name: "14天内", Count: result.RepurchaseUsers},
		{Name: "30天内", Count: 0},
		{Name: "60天内", Count: 0},
		{Name: "60天以上", Count: 0},
	}
	return result, nil
}

func (s *AnalyticsService) ExportTaskID(ctx context.Context) string {
	return fmt.Sprintf("ANA-EXP-%d", time.Now().UnixNano())
}

func orderQuery(ctx context.Context, search carpoolReq.AnalyticsSearch) *gorm.DB {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.OrderMain{})
	if search.ServiceType != "" {
		db = db.Where("service_type = ?", search.ServiceType)
	}
	return db
}

func analyticsRange(search carpoolReq.AnalyticsSearch, defaultStart, defaultEnd time.Time) (time.Time, time.Time, error) {
	start := time.Date(defaultStart.Year(), defaultStart.Month(), defaultStart.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(defaultEnd.Year(), defaultEnd.Month(), defaultEnd.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if search.StartDate != "" {
		parsed, err := time.ParseInLocation("2006-01-02", search.StartDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start = parsed
	}
	if search.EndDate != "" {
		parsed, err := time.ParseInLocation("2006-01-02", search.EndDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end = parsed.AddDate(0, 0, 1)
	}
	return start, end, nil
}

func validOrderStatuses() []string {
	return []string{"paid", "completed"}
}

func invalidOrderStatuses() []string {
	return []string{"cancelled", "canceled", "refunded"}
}

func percent(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return roundAnalytics(float64(part) * 100 / float64(total))
}

func growthRate(previous, current int64) float64 {
	if previous <= 0 {
		return 0
	}
	return percent(current-previous, previous)
}

func roundAnalytics(v float64) float64 {
	return math.Round(v*100) / 100
}
