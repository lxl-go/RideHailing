package server

import (
	"context"
	"math"
	"strconv"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
)

type mobilePassengerProfileService interface {
	EnsurePassenger(context.Context, int64, string) (*passengerv1.PassengerProfileReply, error)
}

type mobileDriverProfileService interface {
	EnsureDriver(context.Context, int64, string) (*driverv1.DriverProfileReply, error)
}

func mobileOrderDetailResponse(reply *orderv1.GetOrderDetailReply) map[string]any {
	order := mobileOrderItemResponse(reply.GetOrder())
	return map[string]any{
		"order": order,
		"id":    order["id"],
	}
}

func mobileOrderListResponse(reply *orderv1.ListOrdersReply) map[string]any {
	items := make([]map[string]any, 0, len(reply.GetItems()))
	for _, item := range reply.GetItems() {
		items = append(items, mobileOrderItemResponse(item))
	}
	return map[string]any{"total": reply.GetTotal(), "items": items, "list": items}
}

func mobilePendingOrderListResponse(reply *orderv1.PendingOrdersReply) map[string]any {
	items := make([]map[string]any, 0, len(reply.GetItems()))
	for _, item := range reply.GetItems() {
		totalPriceText, totalPriceCents := mobileMoneyFields(item.GetTotalPrice())
		items = append(items, map[string]any{
			"id":                int64String(item.GetId()),
			"orderId":           int64String(item.GetId()),
			"tripId":            int64String(item.GetTripId()),
			"trip_id":           int64String(item.GetTripId()),
			"passengerId":       int64String(item.GetPassengerId()),
			"passenger_id":      int64String(item.GetPassengerId()),
			"seats":             item.GetSeatsBooked(),
			"seatsBooked":       item.GetSeatsBooked(),
			"seats_booked":      item.GetSeatsBooked(),
			"totalPrice":        item.GetTotalPrice(),
			"total_price":       item.GetTotalPrice(),
			"totalPriceText":    totalPriceText,
			"total_price_text":  totalPriceText,
			"totalPriceCents":   totalPriceCents,
			"total_price_cents": totalPriceCents,
			"amount":            item.GetTotalPrice(),
			"amountText":        totalPriceText,
			"amount_text":       totalPriceText,
			"amountCents":       totalPriceCents,
			"amount_cents":      totalPriceCents,
			"status":            mobileOrderStatus(item.GetStatus()),
			"rawStatus":         item.GetStatus(),
			"createdAt":         item.GetCreatedAt(),
			"created_at":        item.GetCreatedAt(),
		})
	}
	return map[string]any{"total": reply.GetTotal(), "items": items, "list": items}
}

func mobileOrderItemResponse(item *orderv1.OrderItem) map[string]any {
	if item == nil {
		return map[string]any{}
	}
	totalPriceText, totalPriceCents := mobileMoneyFields(item.GetTotalPrice())
	return map[string]any{
		"id":                int64String(item.GetId()),
		"orderId":           int64String(item.GetId()),
		"tripId":            int64String(item.GetTripId()),
		"trip_id":           int64String(item.GetTripId()),
		"passengerId":       int64String(item.GetPassengerId()),
		"passenger_id":      int64String(item.GetPassengerId()),
		"driverId":          int64String(item.GetDriverId()),
		"driver_id":         int64String(item.GetDriverId()),
		"origin":            item.GetOrigin(),
		"destination":       item.GetDestination(),
		"departTime":        item.GetDepartTime(),
		"depart_time":       item.GetDepartTime(),
		"seats":             item.GetSeatsBooked(),
		"seatsBooked":       item.GetSeatsBooked(),
		"seats_booked":      item.GetSeatsBooked(),
		"totalPrice":        item.GetTotalPrice(),
		"total_price":       item.GetTotalPrice(),
		"totalPriceText":    totalPriceText,
		"total_price_text":  totalPriceText,
		"totalPriceCents":   totalPriceCents,
		"total_price_cents": totalPriceCents,
		"amount":            item.GetTotalPrice(),
		"amountText":        totalPriceText,
		"amount_text":       totalPriceText,
		"amountCents":       totalPriceCents,
		"amount_cents":      totalPriceCents,
		"status":            mobileOrderStatus(item.GetStatus()),
		"rawStatus":         item.GetStatus(),
		"createdAt":         item.GetCreatedAt(),
		"created_at":        item.GetCreatedAt(),
		"acceptedAt":        item.GetAcceptedAt(),
		"accepted_at":       item.GetAcceptedAt(),
		"rejectReason":      item.GetRejectReason(),
		"reject_reason":     item.GetRejectReason(),
		"rejectedAt":        item.GetRejectedAt(),
		"rejected_at":       item.GetRejectedAt(),
		"refundAmount":      item.GetRefundAmount(),
		"refund_amount":     item.GetRefundAmount(),
		"refundedAt":        item.GetRefundedAt(),
		"refunded_at":       item.GetRefundedAt(),
	}
}

func enrichMobileOrderContactPayload(ctx context.Context, payload map[string]any, passengerSvc mobilePassengerProfileService, driverSvc mobileDriverProfileService) {
	if payload == nil {
		return
	}
	order, _ := payload["order"].(map[string]any)
	if order == nil {
		order = payload
	}
	enrichMobileOrderContacts(ctx, order, passengerSvc, driverSvc)
}

func enrichMobileOrderContacts(ctx context.Context, order map[string]any, passengerSvc mobilePassengerProfileService, driverSvc mobileDriverProfileService) {
	if order == nil {
		return
	}
	passengerID, _ := parseMobileInt64Value(firstMapValue(order, "passengerId", "passenger_id"))
	if passengerSvc != nil && passengerID > 0 {
		if reply, err := passengerSvc.EnsurePassenger(ctx, passengerID, ""); err == nil && reply.GetPassenger() != nil {
			passenger := reply.GetPassenger()
			setContactFields(order, "passenger", passenger.GetNickname(), passenger.GetPhone())
		}
	}
	driverID, _ := parseMobileInt64Value(firstMapValue(order, "driverId", "driver_id"))
	if driverSvc != nil && driverID > 0 {
		if reply, err := driverSvc.EnsureDriver(ctx, driverID, ""); err == nil && reply.GetDriver() != nil {
			driver := reply.GetDriver()
			setContactFields(order, "driver", driver.GetName(), driver.GetPhone())
		}
	}
}

func firstMapValue(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return nil
}

func setContactFields(order map[string]any, prefix, name, phone string) {
	if name != "" {
		order[prefix+"Name"] = name
		order[prefix+"_name"] = name
	}
	if phone != "" {
		order[prefix+"Mobile"] = phone
		order[prefix+"_mobile"] = phone
		order[prefix+"Phone"] = phone
		order[prefix+"_phone"] = phone
	}
}

func mobileCreateOrderResponse(reply *orderv1.CreateOrderReply) map[string]any {
	id := int64String(reply.GetOrderId())
	totalPriceText, totalPriceCents := mobileMoneyFields(reply.GetTotalPrice())
	return map[string]any{
		"orderId":           id,
		"order_id":          id,
		"id":                id,
		"totalPrice":        reply.GetTotalPrice(),
		"total_price":       reply.GetTotalPrice(),
		"totalPriceText":    totalPriceText,
		"total_price_text":  totalPriceText,
		"totalPriceCents":   totalPriceCents,
		"total_price_cents": totalPriceCents,
		"amount":            reply.GetTotalPrice(),
		"amountText":        totalPriceText,
		"amount_text":       totalPriceText,
		"amountCents":       totalPriceCents,
		"amount_cents":      totalPriceCents,
	}
}

func mobileMoneyFields(value float64) (string, int64) {
	cents := int64(math.Round(value * 100))
	return strconv.FormatFloat(float64(cents)/100, 'f', 2, 64), cents
}

func mobileDriverIncomeResponse(reply *orderv1.DriverIncomeReply) map[string]any {
	if reply == nil {
		return map[string]any{}
	}
	records := make([]map[string]any, 0, len(reply.GetRecords()))
	for _, item := range reply.GetRecords() {
		amountText, amountCents := mobileMoneyFields(item.GetAmount())
		records = append(records, map[string]any{
			"orderId":      int64String(item.GetOrderId()),
			"order_id":     int64String(item.GetOrderId()),
			"passengerId":  int64String(item.GetPassengerId()),
			"passenger_id": int64String(item.GetPassengerId()),
			"tripId":       int64String(item.GetTripId()),
			"trip_id":      int64String(item.GetTripId()),
			"origin":       item.GetOrigin(),
			"destination":  item.GetDestination(),
			"amount":       item.GetAmount(),
			"amountText":   amountText,
			"amount_text":  amountText,
			"amountCents":  amountCents,
			"amount_cents": amountCents,
			"status":       mobileOrderStatus(item.GetStatus()),
			"rawStatus":    item.GetStatus(),
			"acceptedAt":   item.GetAcceptedAt(),
			"accepted_at":  item.GetAcceptedAt(),
		})
	}
	return map[string]any{
		"todayOrders":      reply.GetTodayOrders(),
		"today_orders":     reply.GetTodayOrders(),
		"todayIncome":      reply.GetTodayIncome(),
		"today_income":     reply.GetTodayIncome(),
		"pendingWithdraw":  reply.GetPendingWithdraw(),
		"pending_withdraw": reply.GetPendingWithdraw(),
		"records":          records,
		"list":             records,
	}
}

func mobileOrderStatus(status int32) string {
	switch status {
	case 0:
		return "pending"
	case 1:
		return "accepted"
	case 2:
		return "completed"
	case 3:
		return "cancelled"
	case 4:
		return "paid"
	case 5:
		return "picking_up"
	case 6:
		return "delivering"
	default:
		return "unknown"
	}
}

func parseMobileOrderStatus(value string) int32 {
	switch value {
	case "":
		return -1
	case "pending", "waiting", "waiting_pay":
		return 0
	case "accepted", "ongoing":
		return 1
	case "completed":
		return 2
	case "cancelled", "canceled":
		return 3
	case "paid":
		return 4
	case "picking_up":
		return 5
	case "delivering":
		return 6
	default:
		parsed := int32(parseInt(value))
		if parsed == 0 && value != "0" {
			return -1
		}
		return parsed
	}
}
