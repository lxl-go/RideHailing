package carpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"gorm.io/gorm"
)

const (
	dispatchActionScore    = "score"
	dispatchActionCancel   = "cancel"
	dispatchActionReassign = "reassign"
	dispatchDecisionSelect = "selected"
	dispatchDecisionCancel = "cancelled"
	dispatchDecisionAssign = "reassigned"
)

var dispatchAuditSeq uint64

type DispatchService struct{}

type DispatchOrderDetail struct {
	Order  carpoolModel.AdminOrderView       `json:"order"`
	Audits []carpoolModel.OrderDispatchAudit `json:"audits"`
}

type DispatchDecision struct {
	AuditNo          string                   `json:"auditNo"`
	OrderID          string                   `json:"orderId"`
	OrderNo          string                   `json:"orderNo"`
	SelectedDriverID string                   `json:"selectedDriverId"`
	SelectedDriver   string                   `json:"selectedDriver"`
	Score            float64                  `json:"score"`
	ScoreDetail      string                   `json:"scoreDetail"`
	DecisionReason   string                   `json:"decisionReason"`
	Candidates       []DispatchCandidateScore `json:"candidates"`
	ExcludedReasons  map[string]string        `json:"excludedReasons"`
}

type DispatchCandidateScore struct {
	DriverID    string  `json:"driverId"`
	DriverName  string  `json:"driverName"`
	Distance    float64 `json:"distance"`
	Rating      float64 `json:"rating"`
	Idle        float64 `json:"idle"`
	Total       float64 `json:"total"`
	VehicleType string  `json:"vehicleType"`
}

type dispatchWeights struct {
	Distance float64 `json:"distance"`
	Rating   float64 `json:"rating"`
	Idle     float64 `json:"idle"`
	Mode     string  `json:"mode"`
}

func (s *DispatchService) ListOrders(ctx context.Context, search carpoolReq.DispatchOrderSearch) ([]carpoolModel.AdminOrderView, int64, error) {
	orderSearch := carpoolReq.OrderSearch{
		PageInfo:    search.PageInfo,
		OrderNo:     search.OrderNo,
		ServiceType: search.ServiceType,
		Status:      search.Status,
	}
	db := dispatchOrderFilterQuery(realOrderSearchQuery(ctx, orderSearch), search)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var rows []adminOrderRow
	rowsQuery := dispatchOrderFilterQuery(realOrderRowsQuery(ctx, orderSearch), search)
	if err := rowsQuery.Order("ct.departure_time ASC, co.id ASC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return adminOrderViewsFromRows(rows), total, nil
}

func (s *DispatchService) GetOrderDetail(ctx context.Context, id uint64) (*DispatchOrderDetail, error) {
	if id == 0 {
		return nil, errors.New("orderId is required")
	}
	order, err := getRealAdminOrderByNo(ctx, strconv.FormatUint(id, 10))
	if err != nil {
		return nil, err
	}
	var audits []carpoolModel.OrderDispatchAudit
	if err := global.GVA_DB.WithContext(ctx).Where("order_id = ?", id).Order("created_at DESC, id DESC").Find(&audits).Error; err != nil {
		return nil, err
	}
	return &DispatchOrderDetail{Order: order, Audits: audits}, nil
}

func (s *DispatchService) CancelOrder(ctx context.Context, req carpoolReq.CancelOrderRequest) error {
	orderID, err := parsePositiveUintString(req.OrderID, "orderId")
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.Reason) == "" {
		return errors.New("reason is required")
	}
	idem := strings.TrimSpace(req.IdempotencyKey)
	if idem == "" {
		return errors.New("idempotencyKey is required")
	}
	var existing carpoolModel.OrderDispatchAudit
	existingResult := global.GVA_DB.WithContext(ctx).Where("idempotency_key = ?", idem).Find(&existing)
	if existingResult.Error != nil {
		return existingResult.Error
	}
	if existingResult.RowsAffected > 0 {
		return nil
	}

	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order carpoolModel.CarpoolOrderRecord
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
		}
		orderNo := strconv.FormatUint(order.ID, 10)
		if err := tx.Model(&carpoolModel.CarpoolOrderRecord{}).Where("id = ?", order.ID).Updates(map[string]any{
			"status": "cancelled",
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&carpoolModel.OrderStatusHistory{
			OrderNo: orderNo, FromStatus: order.Status, ToStatus: "cancelled", Operator: fallback(req.Operator, "dispatch"), Reason: req.Reason,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&carpoolModel.OrderDispatchAudit{
			AuditNo:        nextDispatchAuditNo(),
			OrderID:        order.ID,
			OrderNo:        orderNo,
			Action:         dispatchActionCancel,
			Decision:       dispatchDecisionCancel,
			DecisionReason: req.Reason,
			IdempotencyKey: idem,
			Operator:       fallback(req.Operator, "dispatch"),
			TraceID:        traceIDFromContext(ctx),
		}).Error
	})
}

func (s *DispatchService) ReassignOrder(ctx context.Context, req carpoolReq.ReassignOrderRequest) (*DispatchDecision, error) {
	orderID, err := parsePositiveUintString(req.OrderID, "orderId")
	if err != nil {
		return nil, err
	}
	driverID, err := parsePositiveUintString(req.DriverID, "driverId")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Reason) == "" {
		return nil, errors.New("reason is required")
	}
	idem := strings.TrimSpace(req.IdempotencyKey)
	if idem == "" {
		return nil, errors.New("idempotencyKey is required")
	}
	var existing carpoolModel.OrderDispatchAudit
	existingResult := global.GVA_DB.WithContext(ctx).Where("idempotency_key = ?", idem).Find(&existing)
	if existingResult.Error != nil {
		return nil, existingResult.Error
	}
	if existingResult.RowsAffected > 0 {
		return decisionFromAudit(existing), nil
	}

	var audit carpoolModel.OrderDispatchAudit
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order carpoolModel.CarpoolOrderRecord
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
		}
		if err := tx.Model(&carpoolModel.CarpoolTripRecord{}).Where("id = ?", order.TripID).Updates(map[string]any{
			"driver_id": driverID,
		}).Error; err != nil {
			return err
		}
		orderNo := strconv.FormatUint(order.ID, 10)
		audit = carpoolModel.OrderDispatchAudit{
			AuditNo:        nextDispatchAuditNo(),
			OrderID:        order.ID,
			OrderNo:        orderNo,
			Action:         dispatchActionReassign,
			DriverID:       driverID,
			DriverName:     strings.TrimSpace(req.DriverName),
			Decision:       dispatchDecisionAssign,
			DecisionReason: req.Reason,
			IdempotencyKey: idem,
			Operator:       fallback(req.Operator, "dispatch"),
			TraceID:        traceIDFromContext(ctx),
		}
		return tx.Create(&audit).Error
	})
	if err != nil {
		return nil, err
	}
	return decisionFromAudit(audit), nil
}

func (s *DispatchService) ScoreDrivers(ctx context.Context, req carpoolReq.DispatchScoreRequest) (*DispatchDecision, error) {
	orderID, err := parsePositiveUintString(req.OrderID, "orderId")
	if err != nil {
		return nil, err
	}
	if len(req.Candidates) == 0 {
		return nil, errors.New("candidates is required")
	}
	idem := strings.TrimSpace(req.IdempotencyKey)
	if idem != "" {
		var existing carpoolModel.OrderDispatchAudit
		existingResult := global.GVA_DB.WithContext(ctx).Where("idempotency_key = ?", idem).Find(&existing)
		if existingResult.Error != nil {
			return nil, existingResult.Error
		}
		if existingResult.RowsAffected > 0 {
			return decisionFromAudit(existing), nil
		}
	} else {
		idem = fmt.Sprintf("score:%s:%d", req.OrderID, time.Now().UnixNano())
	}

	order, err := getRealAdminOrderByNo(ctx, strconv.FormatUint(orderID, 10))
	if err != nil {
		return nil, err
	}
	weights := s.weightsFor(ctx, req, order.DepartTime)
	decision, err := scoreDispatchCandidates(order, req, weights)
	if err != nil {
		return nil, err
	}
	decision.AuditNo = nextDispatchAuditNo()
	decision.OrderID = order.ID
	decision.OrderNo = order.OrderNo
	selectedDriverID, err := parsePositiveUintString(decision.SelectedDriverID, "selectedDriverId")
	if err != nil {
		return nil, err
	}
	detailBytes, _ := json.Marshal(map[string]any{
		"weights":    weights,
		"candidates": decision.Candidates,
		"excluded":   decision.ExcludedReasons,
	})
	decision.ScoreDetail = string(detailBytes)

	audit := carpoolModel.OrderDispatchAudit{
		AuditNo:        decision.AuditNo,
		OrderID:        orderID,
		OrderNo:        order.OrderNo,
		Action:         dispatchActionScore,
		DriverID:       selectedDriverID,
		DriverName:     decision.SelectedDriver,
		Decision:       dispatchDecisionSelect,
		DecisionReason: decision.DecisionReason,
		Score:          decision.Score,
		ScoreDetail:    decision.ScoreDetail,
		IdempotencyKey: idem,
		Operator:       fallback(req.Operator, "dispatch"),
		TraceID:        traceIDFromContext(ctx),
	}
	if err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(&audit).Error
	}); err != nil {
		return nil, err
	}
	return decision, nil
}

func (s *DispatchService) ListDispatchAudits(ctx context.Context, search carpoolReq.DispatchAuditSearch) ([]carpoolModel.OrderDispatchAudit, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.OrderDispatchAudit{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("audit_no LIKE ? OR order_no LIKE ? OR decision_reason LIKE ?", keyword, keyword, keyword)
	}
	if search.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+strings.TrimSpace(search.OrderNo)+"%")
	}
	if search.Action != "" {
		db = db.Where("action = ?", strings.TrimSpace(search.Action))
	}
	if search.DriverID != 0 {
		db = db.Where("driver_id = ?", search.DriverID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.OrderDispatchAudit
	if err := db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *DispatchService) ListConfigs(ctx context.Context) ([]carpoolModel.DispatchConfig, error) {
	var list []carpoolModel.DispatchConfig
	err := global.GVA_DB.WithContext(ctx).Order("published DESC, city ASC, id DESC").Find(&list).Error
	return list, err
}

func (s *DispatchService) SaveConfig(ctx context.Context, req carpoolReq.DispatchConfigRequest) (*carpoolModel.DispatchConfig, error) {
	if strings.TrimSpace(req.City) == "" {
		return nil, errors.New("city is required")
	}
	config := &carpoolModel.DispatchConfig{
		ConfigNo:            fallback(req.ConfigNo, nextDispatchConfigNo()),
		City:                strings.TrimSpace(req.City),
		FleetID:             strings.TrimSpace(req.FleetID),
		HotZone:             strings.TrimSpace(req.HotZone),
		DayDistanceWeight:   defaultIfZero(req.DayDistanceWeight, 0.65),
		DayRatingWeight:     defaultIfZero(req.DayRatingWeight, 0.25),
		DayIdleWeight:       defaultIfZero(req.DayIdleWeight, 0.10),
		NightDistanceWeight: defaultIfZero(req.NightDistanceWeight, 0.15),
		NightRatingWeight:   defaultIfZero(req.NightRatingWeight, 0.75),
		NightIdleWeight:     defaultIfZero(req.NightIdleWeight, 0.10),
		Enabled:             req.Enabled,
	}
	if req.ID != 0 {
		config.ID = req.ID
	}
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if config.ID != 0 {
			return tx.Model(&carpoolModel.DispatchConfig{}).Where("id = ?", config.ID).Updates(config).Error
		}
		return tx.Create(config).Error
	})
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *DispatchService) PublishConfig(ctx context.Context, id uint64, operator string) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var config carpoolModel.DispatchConfig
		if err := tx.Where("id = ?", id).First(&config).Error; err != nil {
			return err
		}
		if err := tx.Model(&carpoolModel.DispatchConfig{}).Where("id = ?", id).Updates(map[string]any{"published": true, "version": gorm.Expr("version + ?", 1)}).Error; err != nil {
			return err
		}
		snapshot, _ := json.Marshal(config)
		return tx.Create(&carpoolModel.DispatchConfigVersion{ConfigNo: config.ConfigNo, Version: config.Version + 1, SnapshotJSON: string(snapshot), Operator: operator}).Error
	})
}

func (s *DispatchService) RollbackConfig(ctx context.Context, id uint64, operator string) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(&carpoolModel.DispatchConfig{}).Where("id = ?", id).Updates(map[string]any{"published": false, "version": gorm.Expr("version + ?", 1)}).Error
	})
}

func (s *DispatchService) ReplayTrack(ctx context.Context, search carpoolReq.TrackReplaySearch) ([]carpoolModel.DriverLocationPoint, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.DriverLocationPoint{})
	if strings.TrimSpace(search.DriverID) != "" {
		driverID, err := parsePositiveUintString(search.DriverID, "driverId")
		if err != nil {
			return nil, 0, err
		}
		db = db.Where("driver_id = ?", driverID)
	}
	if search.StartTime != "" {
		db = db.Where("reported_at >= ?", search.StartTime)
	}
	if search.EndTime != "" {
		db = db.Where("reported_at <= ?", search.EndTime)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 100
	}
	var list []carpoolModel.DriverLocationPoint
	if err := db.Order("reported_at ASC, id ASC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *DispatchService) ExportTaskID(ctx context.Context) string {
	return fmt.Sprintf("WO11-DISPATCH-EXPORT-%d", time.Now().UnixNano())
}

func (s *DispatchService) weightsFor(ctx context.Context, req carpoolReq.DispatchScoreRequest, depart time.Time) dispatchWeights {
	var cfg carpoolModel.DispatchConfig
	query := global.GVA_DB.WithContext(ctx).Where("enabled = ? AND published = ?", true, true)
	if req.City != "" {
		query = query.Where("city = ?", strings.TrimSpace(req.City))
	}
	if req.FleetID != "" {
		query = query.Where("(fleet_id = ? OR fleet_id = '')", strings.TrimSpace(req.FleetID))
	}
	if req.HotZone != "" {
		query = query.Where("(hot_zone = ? OR hot_zone = '')", strings.TrimSpace(req.HotZone))
	}
	result := query.Order("hot_zone DESC, fleet_id DESC, id DESC").Find(&cfg)
	if result.Error == nil && result.RowsAffected > 0 {
		if isDispatchDaytime(depart) {
			return normalizeWeights(dispatchWeights{Distance: cfg.DayDistanceWeight, Rating: cfg.DayRatingWeight, Idle: cfg.DayIdleWeight, Mode: "day"})
		}
		return normalizeWeights(dispatchWeights{Distance: cfg.NightDistanceWeight, Rating: cfg.NightRatingWeight, Idle: cfg.NightIdleWeight, Mode: "night"})
	}
	if isDispatchDaytime(depart) {
		return dispatchWeights{Distance: 0.65, Rating: 0.25, Idle: 0.10, Mode: "day"}
	}
	return dispatchWeights{Distance: 0.15, Rating: 0.75, Idle: 0.10, Mode: "night"}
}

func scoreDispatchCandidates(order carpoolModel.AdminOrderView, req carpoolReq.DispatchScoreRequest, weights dispatchWeights) (*DispatchDecision, error) {
	decision := &DispatchDecision{ExcludedReasons: map[string]string{}}
	var best *DispatchCandidateScore
	var bestName string
	var bestNumericDriverID uint64
	for _, candidate := range req.Candidates {
		reason := exclusionReason(order, req, candidate)
		key := strings.TrimSpace(candidate.DriverID)
		if reason != "" {
			decision.ExcludedReasons[key] = reason
			continue
		}
		driverID, err := parsePositiveUintString(candidate.DriverID, "driverId")
		if err != nil {
			decision.ExcludedReasons[key] = err.Error()
			continue
		}
		score := DispatchCandidateScore{
			DriverID:    strconv.FormatUint(driverID, 10),
			DriverName:  strings.TrimSpace(candidate.DriverName),
			Distance:    normalizedDistanceScore(candidate.DistanceKM),
			Rating:      normalizedRatingScore(candidate.Rating),
			Idle:        normalizedIdleScore(candidate.IdleMinutes),
			VehicleType: strings.TrimSpace(candidate.VehicleType),
		}
		score.Total = roundAnalytics(score.Distance*weights.Distance + score.Rating*weights.Rating + score.Idle*weights.Idle)
		decision.Candidates = append(decision.Candidates, score)
		if best == nil || score.Total > best.Total || (score.Total == best.Total && driverID < bestNumericDriverID) {
			copyScore := score
			best = &copyScore
			bestName = score.DriverName
			bestNumericDriverID = driverID
		}
	}
	if best == nil {
		return nil, errors.New("no available driver candidate")
	}
	decision.SelectedDriverID = best.DriverID
	decision.SelectedDriver = bestName
	decision.Score = best.Total
	decision.DecisionReason = fmt.Sprintf("selected by %s weights with total score %.2f", weights.Mode, best.Total)
	return decision, nil
}

func exclusionReason(order carpoolModel.AdminOrderView, req carpoolReq.DispatchScoreRequest, candidate carpoolReq.DriverCandidate) string {
	if strings.TrimSpace(candidate.DriverID) == "" {
		return "driver id is required"
	}
	if !candidate.Online {
		return "driver offline"
	}
	if req.City != "" && strings.TrimSpace(candidate.City) != strings.TrimSpace(req.City) {
		return "city mismatch"
	}
	if req.FleetID != "" && strings.TrimSpace(candidate.FleetID) != strings.TrimSpace(req.FleetID) {
		return "fleet mismatch"
	}
	if req.HotZone != "" && strings.TrimSpace(candidate.HotZone) != strings.TrimSpace(req.HotZone) {
		return "hot zone mismatch"
	}
	for _, window := range candidate.ServiceWindows {
		if window.Start.Before(order.ArrivalTime) && window.End.After(order.DepartTime) {
			return "service window overlaps"
		}
	}
	return ""
}

func decisionFromAudit(audit carpoolModel.OrderDispatchAudit) *DispatchDecision {
	return &DispatchDecision{
		AuditNo:          audit.AuditNo,
		OrderID:          strconv.FormatUint(audit.OrderID, 10),
		OrderNo:          audit.OrderNo,
		SelectedDriverID: strconv.FormatUint(audit.DriverID, 10),
		SelectedDriver:   audit.DriverName,
		Score:            audit.Score,
		ScoreDetail:      audit.ScoreDetail,
		DecisionReason:   audit.DecisionReason,
		ExcludedReasons:  map[string]string{},
	}
}

func isDispatchDaytime(t time.Time) bool {
	hour := t.Hour()
	return hour >= 7 && hour < 22
}

func normalizeWeights(w dispatchWeights) dispatchWeights {
	total := w.Distance + w.Rating + w.Idle
	if total <= 0 {
		if w.Mode == "night" {
			return dispatchWeights{Distance: 0.15, Rating: 0.75, Idle: 0.10, Mode: w.Mode}
		}
		return dispatchWeights{Distance: 0.65, Rating: 0.25, Idle: 0.10, Mode: w.Mode}
	}
	w.Distance = w.Distance / total
	w.Rating = w.Rating / total
	w.Idle = w.Idle / total
	return w
}

func normalizedDistanceScore(distanceKM float64) float64 {
	if distanceKM < 0 {
		distanceKM = 0
	}
	return math.Max(0, 100-distanceKM*10)
}

func normalizedRatingScore(rating float64) float64 {
	if rating < 0 {
		rating = 0
	}
	if rating > 5 {
		rating = 5
	}
	return rating * 20
}

func normalizedIdleScore(idleMinutes float64) float64 {
	if idleMinutes < 0 {
		idleMinutes = 0
	}
	return math.Min(100, idleMinutes*2)
}

func defaultIfZero(value, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

func nextDispatchAuditNo() string {
	seq := atomic.AddUint64(&dispatchAuditSeq, 1)
	return fmt.Sprintf("DISP-AUD-%d-%d", time.Now().UnixNano(), seq)
}

func nextDispatchConfigNo() string {
	return fmt.Sprintf("DISP-CFG-%d", time.Now().UnixNano())
}

func parsePositiveUintString(raw string, field string) (uint64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, errors.New(field + "必须是正整数")
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New(field + "必须是正整数")
	}
	return id, nil
}

func dispatchOrderFilterQuery(db *gorm.DB, search carpoolReq.DispatchOrderSearch) *gorm.DB {
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("CAST(co.id AS CHAR) LIKE ? OR cp.out_trade_no LIKE ?", keyword, keyword)
	}
	if search.Plate != "" || search.Phone != "" {
		db = db.Where("1 = 0")
	}
	return db
}
