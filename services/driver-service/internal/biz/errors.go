package biz

import "errors"

var (
	ErrInvalidDriver         = errors.New("invalid driver")
	ErrDriverNotFound        = errors.New("driver not found")
	ErrCertificationNotFound = errors.New("driver certification not found")
	ErrVehicleNotFound       = errors.New("driver vehicle not found")
	ErrVehiclePlateInUse     = errors.New("driver vehicle plate already in use")
	ErrInvalidDriverLocation = errors.New("invalid driver location")
	ErrRealNameNotMatched    = errors.New("real-name authentication failed")
	ErrRealNameUnavailable   = errors.New("real-name authentication unavailable")
)
