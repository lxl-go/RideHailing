package biz

import "errors"

var (
	ErrInvalidPassenger   = errors.New("invalid passenger")
	ErrPassengerNotFound  = errors.New("passenger not found")
	ErrPassengerDuplicate = errors.New("passenger already exists")
)
