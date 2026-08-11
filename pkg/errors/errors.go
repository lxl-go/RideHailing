package errors

import (
	"fmt"
)

// Kind 定义错误类别（对标文档 4.3 节错误 Kind）
type Kind string

const (
	KindInvalidArgument  Kind = "INVALID_ARGUMENT"
	KindUnauthenticated  Kind = "UNAUTHENTICATED"
	KindPermissionDenied Kind = "PERMISSION_DENIED"
	KindNotFound         Kind = "NOT_FOUND"
	KindConflict         Kind = "CONFLICT"
	KindRateLimit        Kind = "RATE_LIMIT"
	KindInternal         Kind = "INTERNAL"
	KindUnavailable      Kind = "UNAVAILABLE"
)

// Error 是统一业务错误（对标文档错误规范）
type Error struct {
	Kind    Kind   `json:"kind"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

// HTTPStatus 返回 Kind 对应的 HTTP 状态码
func (e *Error) HTTPStatus() int {
	switch e.Kind {
	case KindInvalidArgument:
		return 400
	case KindUnauthenticated:
		return 401
	case KindPermissionDenied:
		return 403
	case KindNotFound:
		return 404
	case KindConflict:
		return 409
	case KindRateLimit:
		return 429
	case KindUnavailable:
		return 503
	default:
		return 500
	}
}

// GRPCCode 返回 Kind 对应的 gRPC 状态码字符串
func (e *Error) GRPCCode() string {
	switch e.Kind {
	case KindInvalidArgument:
		return "INVALID_ARGUMENT"
	case KindUnauthenticated:
		return "UNAUTHENTICATED"
	case KindPermissionDenied:
		return "PERMISSION_DENIED"
	case KindNotFound:
		return "NOT_FOUND"
	case KindConflict:
		return "ALREADY_EXISTS"
	case KindRateLimit:
		return "UNAVAILABLE"
	case KindUnavailable:
		return "UNAVAILABLE"
	default:
		return "INTERNAL"
	}
}

// --- 构造函数（对标文档错误封装方式）---

func InvalidArgument(code, msg string) *Error {
	return &Error{Kind: KindInvalidArgument, Code: code, Message: msg}
}

func Unauthenticated(code, msg string) *Error {
	return &Error{Kind: KindUnauthenticated, Code: code, Message: msg}
}

func PermissionDenied(code, msg string) *Error {
	return &Error{Kind: KindPermissionDenied, Code: code, Message: msg}
}

func NotFound(code, msg string) *Error {
	return &Error{Kind: KindNotFound, Code: code, Message: msg}
}

func Conflict(code, msg string) *Error {
	return &Error{Kind: KindConflict, Code: code, Message: msg}
}

func RateLimit(code, msg string) *Error {
	return &Error{Kind: KindRateLimit, Code: code, Message: msg}
}

func Internal(code, msg string, err error) *Error {
	return &Error{Kind: KindInternal, Code: code, Message: msg, Err: err}
}

func Unavailable(code, msg string) *Error {
	return &Error{Kind: KindUnavailable, Code: code, Message: msg}
}

// Wrap 包装底层 error，支持任意 Kind（对标文档复杂场景）
func Wrap(kind Kind, code, msg string, err error) *Error {
	return &Error{Kind: kind, Code: code, Message: msg, Err: err}
}
