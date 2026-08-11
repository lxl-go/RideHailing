package server

import (
	"net/http"
	"sort"
	"strings"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"ride-hailing/services/gateway-service/internal/data"
	"ride-hailing/services/gateway-service/internal/service"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

func registerMobileCompatibilityRoutes(srv *khttp.Server, tripSvc *service.TripService, reviewSvc *service.ReviewService) {
	router := srv.Route("/")
	router.GET("/carpool/coupons", func(ctx khttp.Context) error {
		query := ctx.Query()
		reply, err := tripSvc.ListCoupons(ctx, &tripv1.ListCouponsRequest{
			UserId:   currentUserID(ctx.Request()),
			Page:     int32(parseInt(query.Get("page"))),
			PageSize: int32(parseInt(query.Get("page_size"))),
		})
		return returnData(ctx, mobileCouponListResponse(reply), err)
	})
	router.POST("/carpool/coupons/claim", func(ctx khttp.Context) error {
		req := new(mobileClaimCouponRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := tripSvc.ClaimCoupon(ctx, &tripv1.ClaimCouponRequest{
			UserId:         currentUserID(ctx.Request()),
			CouponNo:       req.couponNo(),
			IdempotencyKey: idempotencyKeyFromRequest(ctx),
		})
		return returnData(ctx, mobileClaimCouponResponse(reply), err)
	})
	router.GET("/carpool/reviews/mine/{orderId}", func(ctx khttp.Context) error {
		orderID, err := parseOrderIDParam(ctx.Vars().Get("orderId"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		reply, err := reviewSvc.GetMyReview(ctx, &reviewv1.GetMyReviewRequest{
			OrderId:    orderID,
			FromUserId: currentUserID(ctx.Request()),
		})
		return returnData(ctx, mobileMyReviewResponse(reply), err)
	})
	router.POST("/carpool/trips/demands", func(ctx khttp.Context) error {
		req := new(tripv1.PublishDemandRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		req.PassengerId = currentUserID(ctx.Request())
		reply, err := tripSvc.PublishDemand(ctx, req)
		return returnData(ctx, mobileDemandResponse(reply.GetDemand()), err)
	})
	router.GET("/carpool/trips/demands/recommendations", func(ctx khttp.Context) error {
		query := ctx.Query()
		page := int32(parseInt(query.Get("page")))
		pageSize := int32(parseInt(query.Get("page_size")))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 5
		}
		demands, err := tripSvc.ListMyDemands(ctx, &tripv1.ListMyDemandsRequest{
			PassengerId: currentUserID(ctx.Request()),
			Status:      1,
			Page:        1,
			PageSize:    20,
		})
		if err != nil {
			return returnData(ctx, nil, err)
		}
		demand := latestRecommendableDemand(demands.GetItems())
		if demand == nil {
			return returnData(ctx, emptyList(), nil)
		}
		trips, err := tripSvc.SearchTrips(ctx, data.SearchTripsRequest{
			Origin:      demand.GetOrigin(),
			Destination: demand.GetDestination(),
			DepartDate:  departDateOnly(demand.GetDepartTime()),
			Page:        page,
			PageSize:    pageSize,
		})
		return returnData(ctx, mobileRecommendedTripListResponse(trips, demand, time.Now()), err)
	})
	router.GET("/carpool/trips/demands/mine", func(ctx khttp.Context) error {
		query := ctx.Query()
		reply, err := tripSvc.ListMyDemands(ctx, &tripv1.ListMyDemandsRequest{
			PassengerId: currentUserID(ctx.Request()),
			Status:      int32(parseInt(query.Get("status"))),
			Page:        int32(parseInt(query.Get("page"))),
			PageSize:    int32(parseInt(query.Get("page_size"))),
		})
		return returnData(ctx, mobileDemandListResponse(reply), err)
	})
	router.POST("/carpool/trips/demands/{id}/cancel", func(ctx khttp.Context) error {
		demandID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil || demandID <= 0 {
			return returnBadRequest(ctx, "需求ID格式不正确，请刷新后重试")
		}
		err = tripSvc.CancelDemand(ctx, &tripv1.CancelDemandRequest{Id: demandID, PassengerId: currentUserID(ctx.Request())})
		return returnMessage(ctx, "cancelled", err)
	})
	router.DELETE("/carpool/trips/{id}", func(ctx khttp.Context) error {
		tripID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil || tripID <= 0 {
			return returnBadRequest(ctx, "行程ID格式不正确，请刷新后重试")
		}
		err = tripSvc.DeleteTrip(ctx, &tripv1.DeleteTripRequest{Id: tripID, DriverId: currentUserID(ctx.Request())})
		return returnMessage(ctx, "deleted", err)
	})
}

type mobileClaimCouponRequest struct {
	CouponNo      string `json:"coupon_no"`
	CouponNoCamel string `json:"couponNo"`
	CouponID      string `json:"coupon_id"`
	CouponIDCamel string `json:"couponId"`
}

func (r *mobileClaimCouponRequest) couponNo() string {
	for _, value := range []string{r.CouponNo, r.CouponNoCamel, r.CouponID, r.CouponIDCamel} {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mobileCouponListResponse(reply *tripv1.ListCouponsReply) map[string]any {
	if reply == nil {
		return emptyList()
	}
	items := make([]any, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, mobileCouponItem(item))
	}
	return map[string]any{"items": items, "list": items, "total": reply.Total}
}

func mobileClaimCouponResponse(reply *tripv1.ClaimCouponReply) map[string]any {
	if reply == nil {
		return map[string]any{"coupon": nil, "duplicated": false}
	}
	return map[string]any{"coupon": mobileCouponItem(reply.Coupon), "duplicated": reply.Duplicated}
}

func mobileCouponItem(item *tripv1.CouponItem) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":              int64String(item.Id),
		"couponNo":        item.CouponNo,
		"coupon_no":       item.CouponNo,
		"couponCode":      item.CouponCode,
		"coupon_code":     item.CouponCode,
		"name":            item.Name,
		"title":           item.Name,
		"couponType":      item.CouponType,
		"coupon_type":     item.CouponType,
		"amount":          item.FaceValue,
		"discount":        item.DiscountRate,
		"thresholdAmount": item.ThresholdAmount,
		"validFrom":       item.ValidFrom,
		"validTo":         item.ValidTo,
		"status":          item.Status,
		"claimed":         item.Claimed,
	}
}

func mobileDemandListResponse(reply *tripv1.ListMyDemandsReply) map[string]any {
	if reply == nil {
		return emptyList()
	}
	items := make([]any, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, mobileDemandResponse(item))
	}
	return map[string]any{"items": items, "list": items, "total": reply.Total}
}

func mobileDemandResponse(item *tripv1.TripDemandItem) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":           int64String(item.Id),
		"passengerId":  int64String(item.PassengerId),
		"passenger_id": int64String(item.PassengerId),
		"origin":       item.Origin,
		"destination":  item.Destination,
		"departTime":   item.DepartTime,
		"depart_time":  item.DepartTime,
		"seats":        item.Seats,
		"budget":       item.Budget,
		"remark":       item.Remark,
		"status":       item.Status,
		"createdAt":    item.CreatedAt,
		"created_at":   item.CreatedAt,
	}
}

func latestRecommendableDemand(items []*tripv1.TripDemandItem) *tripv1.TripDemandItem {
	var selected *tripv1.TripDemandItem
	var selectedTime time.Time
	for _, item := range items {
		if item == nil || item.GetStatus() != 1 {
			continue
		}
		itemTime, ok := parseMobileTime(item.GetDepartTime())
		if selected == nil || (ok && itemTime.After(selectedTime)) {
			selected = item
			selectedTime = itemTime
		}
	}
	return selected
}

func mobileRecommendedTripListResponse(reply *tripv1.SearchTripsReply, demand *tripv1.TripDemandItem, now time.Time) map[string]any {
	if reply == nil || demand == nil {
		return emptyList()
	}
	type recommendedTrip struct {
		item   *tripv1.TripItem
		score  float64
		reason string
	}
	recommended := make([]recommendedTrip, 0, len(reply.GetItems()))
	for _, item := range reply.GetItems() {
		if !isRecommendableTrip(item, demand, now) {
			continue
		}
		score, reason := tripMatchScore(demand, item)
		recommended = append(recommended, recommendedTrip{item: item, score: score, reason: reason})
	}
	sort.SliceStable(recommended, func(i, j int) bool {
		if recommended[i].score != recommended[j].score {
			return recommended[i].score > recommended[j].score
		}
		leftTime, leftOK := parseMobileTime(recommended[i].item.GetDepartTime())
		rightTime, rightOK := parseMobileTime(recommended[j].item.GetDepartTime())
		if leftOK && rightOK && !leftTime.Equal(rightTime) {
			return leftTime.Before(rightTime)
		}
		if recommended[i].item.GetPrice() != recommended[j].item.GetPrice() {
			return recommended[i].item.GetPrice() < recommended[j].item.GetPrice()
		}
		return recommended[i].item.GetId() < recommended[j].item.GetId()
	})
	items := make([]map[string]any, 0, len(recommended))
	for _, trip := range recommended {
		item := mobileTripItemResponse(trip.item)
		item["matchScore"] = trip.score
		item["match_score"] = trip.score
		item["matchReason"] = trip.reason
		item["match_reason"] = trip.reason
		items = append(items, item)
	}
	return map[string]any{"total": len(items), "items": items, "list": items}
}

func isRecommendableTrip(item *tripv1.TripItem, demand *tripv1.TripDemandItem, now time.Time) bool {
	if item == nil || item.GetStatus() != 1 {
		return false
	}
	if demand.GetSeats() > 0 && item.GetSeatsAvailable() < demand.GetSeats() {
		return false
	}
	departTime, ok := parseMobileTime(item.GetDepartTime())
	return !ok || departTime.After(now)
}

func tripMatchScore(demand *tripv1.TripDemandItem, trip *tripv1.TripItem) (float64, string) {
	score := 0.0
	reasons := make([]string, 0, 4)
	if stringExactOrContains(demand.GetOrigin(), trip.GetOrigin()) {
		score += 40
		reasons = append(reasons, "origin")
	}
	if stringExactOrContains(demand.GetDestination(), trip.GetDestination()) {
		score += 40
		reasons = append(reasons, "destination")
	}
	if demandTime, ok := parseMobileTime(demand.GetDepartTime()); ok {
		if tripTime, tripOK := parseMobileTime(trip.GetDepartTime()); tripOK {
			diff := tripTime.Sub(demandTime)
			if diff < 0 {
				diff = -diff
			}
			switch {
			case diff <= 30*time.Minute:
				score += 15
				reasons = append(reasons, "time")
			case diff <= 2*time.Hour:
				score += 10
				reasons = append(reasons, "time")
			case sameDate(demandTime, tripTime):
				score += 5
			}
		}
	}
	if demand.GetBudget() > 0 && trip.GetPrice() > 0 {
		priceDiff := trip.GetPrice() - demand.GetBudget()
		if priceDiff < 0 {
			priceDiff = -priceDiff
		}
		if priceDiff <= demand.GetBudget()*0.1 {
			score += 10
			reasons = append(reasons, "price")
		} else if trip.GetPrice() <= demand.GetBudget() {
			score += 6
		}
	}
	if demand.GetSeats() > 0 && trip.GetSeatsAvailable() >= demand.GetSeats() {
		score += 5
	}
	if len(reasons) == 0 {
		return score, "basic"
	}
	return score, strings.Join(reasons, ",")
}

func stringExactOrContains(left, right string) bool {
	left = strings.TrimSpace(strings.ToLower(left))
	right = strings.TrimSpace(strings.ToLower(right))
	return left != "" && right != "" && (left == right || strings.Contains(left, right) || strings.Contains(right, left))
}

func departDateOnly(value string) string {
	if t, ok := parseMobileTime(value); ok {
		return t.Format("2006-01-02")
	}
	value = strings.TrimSpace(value)
	if len(value) >= len("2006-01-02") {
		return value[:len("2006-01-02")]
	}
	return ""
}

func parseMobileTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func sameDate(left, right time.Time) bool {
	leftYear, leftMonth, leftDay := left.Date()
	rightYear, rightMonth, rightDay := right.Date()
	return leftYear == rightYear && leftMonth == rightMonth && leftDay == rightDay
}

func mobileMyReviewResponse(reply *reviewv1.GetMyReviewReply) map[string]any {
	if reply == nil || !reply.HasReview {
		return map[string]any{"hasReview": false, "has_review": false, "review": nil}
	}
	review := reply.Review
	return map[string]any{
		"hasReview":  true,
		"has_review": true,
		"review": map[string]any{
			"id":           int64String(review.Id),
			"orderId":      int64String(review.OrderId),
			"order_id":     int64String(review.OrderId),
			"fromUserId":   int64String(review.FromUserId),
			"from_user_id": int64String(review.FromUserId),
			"toUserId":     int64String(review.ToUserId),
			"to_user_id":   int64String(review.ToUserId),
			"rating":       review.Rating,
			"content":      review.Content,
			"createdAt":    review.CreatedAt,
			"created_at":   review.CreatedAt,
		},
	}
}

func emptyList() map[string]any {
	return map[string]any{"list": []any{}, "items": []any{}, "total": 0}
}

func isServiceUnavailablePayload(payload map[string]any) bool {
	code, ok := payload["code"].(float64)
	return ok && int(code) == http.StatusServiceUnavailable
}
