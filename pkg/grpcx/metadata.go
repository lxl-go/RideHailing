package grpcx

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	apperrors "ride-hailing/pkg/errors"

	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	HeaderUserID      = "x-user-id"
	HeaderRequestID   = "x-request-id"
	HeaderTraceparent = "traceparent"
	DefaultTimeout    = 3 * time.Second
)

// UnaryClientInterceptor 自动将 context 中的 x-user-id / x-request-id 注入到 gRPC outgoing metadata。
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		pairs := []string{}
		if uid, ok := ctx.Value(HeaderUserID).(string); ok && uid != "" {
			pairs = append(pairs, HeaderUserID, uid)
		} else if uIDVal := ctx.Value("user_id"); uIDVal != nil {
			if id, ok := uIDVal.(int64); ok && id > 0 {
				pairs = append(pairs, HeaderUserID, strconv.FormatInt(id, 10))
			}
		}
		if reqID, ok := ctx.Value(HeaderRequestID).(string); ok && reqID != "" {
			pairs = append(pairs, HeaderRequestID, reqID)
		}
		if traceparent, ok := ctx.Value(HeaderTraceparent).(string); ok && traceparent != "" {
			pairs = append(pairs, HeaderTraceparent, traceparent)
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerInterceptor 自动从 gRPC incoming metadata 提取 x-user-id，注入 context value。
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = ContextWithIncomingMetadata(ctx)
		start := time.Now()
		reply, err := handler(ctx, req)
		mappedErr := ToStatusError(err)
		logUnaryServerCall(ctx, info.FullMethod, start, mappedErr)
		return reply, mappedErr
	}
}

// ContextWithIncomingMetadata copies standard incoming gRPC metadata into context values.
func ContextWithIncomingMetadata(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(HeaderUserID); len(vals) > 0 && vals[0] != "" {
			if id, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
				ctx = context.WithValue(ctx, "user_id", id)
			}
		}
		if vals := md.Get(HeaderRequestID); len(vals) > 0 && vals[0] != "" {
			ctx = context.WithValue(ctx, HeaderRequestID, vals[0])
		}
		if vals := md.Get(HeaderTraceparent); len(vals) > 0 && vals[0] != "" {
			ctx = context.WithValue(ctx, HeaderTraceparent, vals[0])
		}
	}
	return ctx
}

// UserIDFromCtx 从 context 中提取 user_id
func UserIDFromCtx(ctx context.Context) int64 {
	if id, ok := ctx.Value("user_id").(int64); ok {
		return id
	}
	return 0
}

// RequestIDFromCtx 从 context 中提取 request_id
func RequestIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(HeaderRequestID).(string); ok {
		return id
	}
	return ""
}

func TraceparentFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(HeaderTraceparent).(string); ok {
		return id
	}
	return ""
}

// OutgoingContext 手动将标准头注入到 gRPC outgoing context（用于非拦截器场景）
func OutgoingContext(ctx context.Context) context.Context {
	pairs := []string{}
	if uid, ok := ctx.Value(HeaderUserID).(string); ok && uid != "" {
		pairs = append(pairs, HeaderUserID, uid)
	}
	if reqID, ok := ctx.Value(HeaderRequestID).(string); ok && reqID != "" {
		pairs = append(pairs, HeaderRequestID, reqID)
	}
	if traceparent, ok := ctx.Value(HeaderTraceparent).(string); ok && traceparent != "" {
		pairs = append(pairs, HeaderTraceparent, traceparent)
	}
	if len(pairs) > 0 {
		return metadata.AppendToOutgoingContext(ctx, pairs...)
	}
	return ctx
}

type ClientOptions struct {
	Timeout      time.Duration
	Retry        RetryPolicy
	Logger       *zap.Logger
	LogCalls     bool
	DisableLog   bool
	DisableRetry bool
}

type RetryPolicy struct {
	MaxAttempts       int
	Backoff           time.Duration
	IdempotentMethods map[string]bool
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxAttempts: 2}
}

func ContextWithClientTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

func DialInsecure(ctx context.Context, endpoint string, discovery registry.Discovery, opts ClientOptions) (*grpc.ClientConn, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := ContextWithClientTimeout(ctx, timeout)
	defer cancel()

	interceptors := []grpc.UnaryClientInterceptor{UnaryClientInterceptor()}
	if !opts.DisableLog || opts.LogCalls || opts.Logger != nil {
		interceptors = append(interceptors, TraceLoggingUnaryClientInterceptor(opts.Logger))
	}
	if !opts.DisableRetry {
		interceptors = append(interceptors, RetryUnaryClientInterceptor(opts.Retry))
	}

	dialOptions := []kgrpc.ClientOption{
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(timeout),
		kgrpc.WithUnaryInterceptor(chainUnaryClientInterceptors(interceptors...)),
	}
	if discovery != nil {
		dialOptions = append(dialOptions, kgrpc.WithDiscovery(discovery))
	}
	return kgrpc.DialInsecure(ctx, dialOptions...)
}

func CodeFromError(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if code := status.Code(err); code != codes.Unknown {
		return code
	}

	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case apperrors.KindInvalidArgument:
			return codes.InvalidArgument
		case apperrors.KindUnauthenticated:
			return codes.Unauthenticated
		case apperrors.KindPermissionDenied:
			return codes.PermissionDenied
		case apperrors.KindNotFound:
			return codes.NotFound
		case apperrors.KindConflict:
			return codes.AlreadyExists
		case apperrors.KindRateLimit, apperrors.KindUnavailable:
			return codes.Unavailable
		default:
			return codes.Internal
		}
	}
	return codes.Internal
}

func ToStatusError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}

	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		return status.Error(CodeFromError(err), appErr.Message)
	}
	return status.Error(codes.Internal, "internal error")
}

func RetryUnaryClientInterceptor(policy RetryPolicy) grpc.UnaryClientInterceptor {
	if policy.MaxAttempts <= 0 {
		defaultPolicy := DefaultRetryPolicy()
		policy.MaxAttempts = defaultPolicy.MaxAttempts
		policy.Backoff = defaultPolicy.Backoff
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if !shouldRetry(policy, method, err, attempt) {
				return err
			}
			if policy.Backoff > 0 {
				timer := time.NewTimer(policy.Backoff)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
			}
		}
		return err
	}
}

func TraceLoggingUnaryClientInterceptor(log *zap.Logger) grpc.UnaryClientInterceptor {
	if log == nil {
		log = zap.L()
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		code := CodeFromError(err)

		fields := []zap.Field{
			zap.String("grpc_method", method),
			zap.String("grpc_code", code.String()),
			zap.Int64("duration_ms", time.Since(start).Milliseconds()),
		}
		if reqID := RequestIDFromCtx(ctx); reqID != "" {
			fields = append(fields, zap.String("request_id", reqID))
		}
		if traceparent := TraceparentFromCtx(ctx); traceparent != "" {
			fields = append(fields, zap.String("traceparent", traceparent))
		}
		if userID := UserIDFromCtx(ctx); userID > 0 {
			fields = append(fields, zap.Int64("user_id", userID))
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Warn("grpc client call", fields...)
			return err
		}
		log.Info("grpc client call", fields...)
		return nil
	}
}

func logUnaryServerCall(ctx context.Context, method string, start time.Time, err error) {
	log := zap.L()
	code := CodeFromError(err)
	fields := []zap.Field{
		zap.String("grpc_method", method),
		zap.String("grpc_code", code.String()),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	}
	if reqID := RequestIDFromCtx(ctx); reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if traceparent := TraceparentFromCtx(ctx); traceparent != "" {
		fields = append(fields, zap.String("traceparent", traceparent))
	}
	if userID := UserIDFromCtx(ctx); userID > 0 {
		fields = append(fields, zap.Int64("user_id", userID))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		log.Warn("grpc server call", fields...)
		return
	}
	log.Info("grpc server call", fields...)
}

func shouldRetry(policy RetryPolicy, method string, err error, attempt int) bool {
	if err == nil || attempt >= policy.MaxAttempts {
		return false
	}
	if !isIdempotentMethod(policy, method) {
		return false
	}
	switch CodeFromError(err) {
	case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

func isIdempotentMethod(policy RetryPolicy, method string) bool {
	if policy.IdempotentMethods != nil {
		return policy.IdempotentMethods[method]
	}
	name := method
	if index := strings.LastIndex(method, "/"); index >= 0 {
		name = method[index+1:]
	}
	return strings.HasPrefix(name, "Get") ||
		strings.HasPrefix(name, "List") ||
		strings.HasPrefix(name, "Search") ||
		strings.HasPrefix(name, "Check") ||
		strings.HasPrefix(name, "Verify")
}

func chainUnaryClientInterceptors(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if len(interceptors) == 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		chained := invoker
		for i := len(interceptors) - 1; i >= 0; i-- {
			current := interceptors[i]
			next := chained
			chained = func(current grpc.UnaryClientInterceptor, next grpc.UnaryInvoker) grpc.UnaryInvoker {
				return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
					return current(ctx, method, req, reply, cc, next, opts...)
				}
			}(current, next)
		}
		return chained(ctx, method, req, reply, cc, opts...)
	}
}
