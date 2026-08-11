package server

import tripv1 "ride-hailing/services/trip-service/api/trip/v1"

func mobileSearchTripListResponse(reply *tripv1.SearchTripsReply) map[string]any {
	if reply == nil {
		return emptyList()
	}
	items := make([]map[string]any, 0, len(reply.GetItems()))
	for _, item := range reply.GetItems() {
		items = append(items, mobileTripItemResponse(item))
	}
	return map[string]any{"total": reply.GetTotal(), "items": items, "list": items}
}

func mobileDriverTripListResponse(reply *tripv1.ListDriverTripsReply) map[string]any {
	if reply == nil {
		return emptyList()
	}
	items := make([]map[string]any, 0, len(reply.GetItems()))
	for _, item := range reply.GetItems() {
		items = append(items, mobileTripItemResponse(item))
	}
	return map[string]any{"total": reply.GetTotal(), "items": items, "list": items}
}

func mobileTripDetailResponse(reply *tripv1.GetTripDetailReply) map[string]any {
	trip := mobileTripItemResponse(reply.GetTrip())
	return map[string]any{"trip": trip, "id": trip["id"]}
}

func mobilePublishTripResponse(reply *tripv1.PublishTripReply) map[string]any {
	id := int64String(reply.GetTripId())
	return map[string]any{
		"tripId":      id,
		"trip_id":     id,
		"id":          id,
		"status":      reply.GetStatus(),
		"price":       reply.GetPrice(),
		"arriveTime":  reply.GetArriveTime(),
		"arrive_time": reply.GetArriveTime(),
	}
}

func mobileTripItemResponse(item *tripv1.TripItem) map[string]any {
	if item == nil {
		return map[string]any{}
	}
	driverID := int64String(item.GetDriverId())
	auditOperatorID := int64String(item.GetAuditOperatorId())
	return map[string]any{
		"id":                     int64String(item.GetId()),
		"driverId":               driverID,
		"driver_id":              driverID,
		"origin":                 item.GetOrigin(),
		"destination":            item.GetDestination(),
		"departTime":             item.GetDepartTime(),
		"depart_time":            item.GetDepartTime(),
		"arriveTime":             item.GetArriveTime(),
		"arrive_time":            item.GetArriveTime(),
		"seatsTotal":             item.GetSeatsTotal(),
		"seats_total":            item.GetSeatsTotal(),
		"seatsAvailable":         item.GetSeatsAvailable(),
		"seats_available":        item.GetSeatsAvailable(),
		"price":                  item.GetPrice(),
		"status":                 item.GetStatus(),
		"createdAt":              item.GetCreatedAt(),
		"created_at":             item.GetCreatedAt(),
		"rejectReason":           item.GetRejectReason(),
		"reject_reason":          item.GetRejectReason(),
		"auditOperatorId":        auditOperatorID,
		"audit_operator_id":      auditOperatorID,
		"auditTime":              item.GetAuditTime(),
		"audit_time":             item.GetAuditTime(),
		"routeDistanceMeters":    item.GetRouteDistanceMeters(),
		"route_distance_meters":  item.GetRouteDistanceMeters(),
		"routeDurationSeconds":   item.GetRouteDurationSeconds(),
		"route_duration_seconds": item.GetRouteDurationSeconds(),
		"isDeleted":              item.GetIsDeleted(),
		"is_deleted":             item.GetIsDeleted(),
	}
}
