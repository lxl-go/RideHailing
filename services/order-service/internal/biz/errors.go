package biz

import "errors"

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrTripNotFound          = errors.New("trip not found")
	ErrInvalidOrder          = errors.New("invalid order")
	ErrNotOrderOwner         = errors.New("not the owner of this order")
	ErrNotTripOwner          = errors.New("not the owner of this trip")
	ErrOrderCannotCancel     = errors.New("order cannot be cancelled in current status")
	ErrOrderCannotComplete   = errors.New("order cannot be completed in current status")
	ErrOrderAlreadyHandled   = errors.New("order already handled")
	ErrTripNotAvailable      = errors.New("trip is not available")
	ErrInsufficientSeats     = errors.New("insufficient seats")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrInvalidPayment        = errors.New("invalid payment")
	ErrPaymentAmountMismatch = errors.New("payment amount mismatch")
	ErrPaymentNotSuccessful  = errors.New("payment trade status is not successful")
	ErrRejectReasonRequired  = errors.New("reject reason is required")
)
