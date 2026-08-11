package biz

import "errors"

var ErrTripNotFound = errors.New("trip not found")
var ErrInvalidTrip = errors.New("invalid trip")
var ErrInvalidCoupon = errors.New("invalid coupon")
var ErrCouponNotFound = errors.New("coupon not found")
var ErrCouponStockExhausted = errors.New("coupon stock exhausted")
var ErrInvalidDemand = errors.New("invalid demand")
var ErrDemandNotFound = errors.New("trip demand not found")
var ErrDemandCannotCancel = errors.New("trip demand cannot cancel")
var ErrTripHasActiveOrders = errors.New("trip has active orders")
var ErrTripTimeConflict = errors.New("trip departure conflicts with an approved trip")
var ErrTripNotPending = errors.New("trip is not pending review")
var ErrInvalidReview = errors.New("invalid trip review")
var ErrRedisUnavailable = errors.New("Redis服务不可用，请稍后重试")
var ErrDuplicateTripRequest = errors.New("行程发布请求正在处理中，请稍后查询结果")
