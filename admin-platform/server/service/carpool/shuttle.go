package carpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

type ShuttleService struct{}

func (s *ShuttleService) ListLines(ctx context.Context, search carpoolReq.ShuttleLineSearch) ([]carpoolModel.ShuttleLine, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.ShuttleLine{})
	if search.Status != nil {
		db = db.Where("status = ?", *search.Status)
	}
	if kw := strings.TrimSpace(search.Keyword); kw != "" {
		like := "%" + kw + "%"
		db = db.Where("line_code LIKE ? OR line_name LIKE ? OR route LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.ShuttleLine
	if err := db.Order("sort_no ASC, created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *ShuttleService) GetLine(ctx context.Context, id uint64) (*carpoolModel.ShuttleLine, error) {
	var line carpoolModel.ShuttleLine
	if err := global.GVA_DB.WithContext(ctx).First(&line, id).Error; err != nil {
		return nil, err
	}
	return &line, nil
}

func (s *ShuttleService) CreateLine(ctx context.Context, payload carpoolReq.ShuttleLinePayload) (*carpoolModel.ShuttleLine, error) {
	line, err := payloadToLine(payload)
	if err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&line).Error; err != nil {
		return nil, err
	}
	return &line, nil
}

func (s *ShuttleService) UpdateLine(ctx context.Context, id uint64, payload carpoolReq.ShuttleLinePayload) (*carpoolModel.ShuttleLine, error) {
	line, err := payloadToLine(payload)
	if err != nil {
		return nil, err
	}
	line.ID = id
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.ShuttleLine{}).Where("id = ?", id).Updates(map[string]any{
		"line_code":    line.LineCode,
		"line_name":    line.LineName,
		"route":        line.Route,
		"depart_time":  line.DepartTime,
		"arrive_time":  line.ArriveTime,
		"vehicle_type": line.VehicleType,
		"total_seats":  line.TotalSeats,
		"remain_seats": line.RemainSeats,
		"sort_no":      line.SortNo,
		"stations":     line.Stations,
		"notice":       line.Notice,
	}).Error; err != nil {
		return nil, err
	}
	return s.GetLine(ctx, id)
}

func (s *ShuttleService) PublishLines(ctx context.Context, ids []uint64) error {
	return global.GVA_DB.WithContext(ctx).Model(&carpoolModel.ShuttleLine{}).Where("id IN ?", ids).Update("status", 1).Error
}

func (s *ShuttleService) ExportLines(ctx context.Context) string {
	return fmt.Sprintf("EXP-%d", time.Now().UnixNano())
}

func payloadToLine(payload carpoolReq.ShuttleLinePayload) (carpoolModel.ShuttleLine, error) {
	stations, err := normalizeShuttleStations(payload.Stations)
	if err != nil {
		return carpoolModel.ShuttleLine{}, err
	}
	stationsJSON, err := json.Marshal(stations)
	if err != nil {
		return carpoolModel.ShuttleLine{}, err
	}
	return carpoolModel.ShuttleLine{
		LineCode:    strings.TrimSpace(payload.LineCode),
		LineName:    strings.TrimSpace(payload.LineName),
		Route:       strings.TrimSpace(payload.Route),
		DepartTime:  strings.TrimSpace(payload.DepartTime),
		ArriveTime:  strings.TrimSpace(payload.ArriveTime),
		VehicleType: strings.TrimSpace(payload.VehicleType),
		TotalSeats:  payload.TotalSeats,
		RemainSeats: payload.RemainSeats,
		SortNo:      payload.SortNo,
		Status:      0,
		Stations:    string(stationsJSON),
		Notice:      strings.TrimSpace(payload.Notice),
	}, nil
}

func normalizeShuttleStations(stations []carpoolReq.ShuttleStation) ([]carpoolReq.ShuttleStation, error) {
	if len(stations) == 0 {
		return []carpoolReq.ShuttleStation{}, nil
	}
	byName := map[string]carpoolReq.ShuttleStation{}
	order := make([]string, 0, len(stations))
	for _, station := range stations {
		station.Name = strings.TrimSpace(station.Name)
		station.Time = strings.TrimSpace(station.Time)
		station.Type = strings.TrimSpace(station.Type)
		if station.Name == "" {
			return nil, errors.New("station name is required")
		}
		if station.OffsetMin < 0 {
			return nil, errors.New("station offsetMin cannot be negative")
		}
		key := strings.ToLower(station.Name)
		existing, ok := byName[key]
		if !ok {
			byName[key] = station
			order = append(order, key)
			continue
		}
		if station.OffsetMin < existing.OffsetMin {
			byName[key] = station
		}
	}
	normalized := make([]carpoolReq.ShuttleStation, 0, len(byName))
	for _, key := range order {
		station, ok := byName[key]
		if ok {
			normalized = append(normalized, station)
		}
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].OffsetMin == normalized[j].OffsetMin {
			return normalized[i].Name < normalized[j].Name
		}
		return normalized[i].OffsetMin < normalized[j].OffsetMin
	})
	if normalized[0].OffsetMin != 0 {
		return nil, errors.New("first station offsetMin must be 0")
	}
	return normalized, nil
}
