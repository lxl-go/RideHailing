package biz

import "errors"

var (
	ErrInvalidPrincipal   = errors.New("invalid principal")
	ErrInvalidRole        = errors.New("invalid role")
	ErrAccountNotFound    = errors.New("account not found")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrInvalidToken       = errors.New("invalid token")
	ErrSMSCodeNotFound    = errors.New("sms code not found")
	ErrInvalidSMSCode     = errors.New("invalid sms code")
	ErrSMSCodeTooFrequent = errors.New("please do not request sms code frequently")
	ErrSMSLoginLocked     = errors.New("sms login locked")
	ErrSMSSendFailed      = errors.New("sms send failed")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRevoked     = errors.New("session revoked")
)
