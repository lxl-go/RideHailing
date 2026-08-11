package grpcx

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "ride-hailing/pkg/errors"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryClientInterceptorPropagatesStandardMetadata(t *testing.T) {
	ctx := context.WithValue(context.Background(), HeaderUserID, "1001")
	ctx = context.WithValue(ctx, HeaderRequestID, "req-1")
	ctx = context.WithValue(ctx, HeaderTraceparent, "00-abc-def-01")

	err := UnaryClientInterceptor()(ctx, "/svc.Method", nil, nil, nil, func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		require.Equal(t, []string{"1001"}, md.Get(HeaderUserID))
		require.Equal(t, []string{"req-1"}, md.Get(HeaderRequestID))
		require.Equal(t, []string{"00-abc-def-01"}, md.Get(HeaderTraceparent))
		return nil
	})

	require.NoError(t, err)
}

func TestContextWithClientTimeoutAddsDeadlineWhenMissing(t *testing.T) {
	ctx, cancel := ContextWithClientTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	deadline, ok := ctx.Deadline()
	require.True(t, ok)
	require.WithinDuration(t, time.Now().Add(150*time.Millisecond), deadline, 50*time.Millisecond)
}

func TestContextWithClientTimeoutKeepsEarlierDeadline(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer parentCancel()

	ctx, cancel := ContextWithClientTimeout(parent, time.Second)
	defer cancel()

	parentDeadline, parentOK := parent.Deadline()
	deadline, ok := ctx.Deadline()
	require.True(t, parentOK)
	require.True(t, ok)
	require.Equal(t, parentDeadline, deadline)
}

func TestCodeFromErrorMapsProjectErrors(t *testing.T) {
	err := apperrors.PermissionDenied("AUTHZ_RESOURCE_OWNER_DENIED", "forbidden")

	require.Equal(t, codes.PermissionDenied, CodeFromError(err))
	require.Equal(t, codes.OK, CodeFromError(nil))
	require.Equal(t, codes.NotFound, CodeFromError(status.Error(codes.NotFound, "missing")))
	require.Equal(t, codes.Internal, CodeFromError(errors.New("boom")))
}

func TestRetryUnaryClientInterceptorRetriesOnlyIdempotentUnavailable(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts: 3,
		IdempotentMethods: map[string]bool{
			"/trip.v1.Trip/SearchTrips": true,
		},
	}
	attempts := 0

	err := RetryUnaryClientInterceptor(policy)(context.Background(), "/trip.v1.Trip/SearchTrips", nil, nil, nil, func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "downstream unavailable")
		}
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 3, attempts)
}

func TestRetryUnaryClientInterceptorDefaultRetriesReadLikeMethods(t *testing.T) {
	attempts := 0

	err := RetryUnaryClientInterceptor(RetryPolicy{})(context.Background(), "/trip.v1.Trip/SearchTrips", nil, nil, nil, func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		attempts++
		if attempts < 2 {
			return status.Error(codes.Unavailable, "downstream unavailable")
		}
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 2, attempts)
}

func TestRetryUnaryClientInterceptorDefaultDoesNotRetryWriteLikeMethods(t *testing.T) {
	attempts := 0

	err := RetryUnaryClientInterceptor(RetryPolicy{})(context.Background(), "/order.v1.Order/CreateOrder", nil, nil, nil, func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		attempts++
		return status.Error(codes.Unavailable, "downstream unavailable")
	})

	require.Error(t, err)
	require.Equal(t, 1, attempts)
}

func TestRetryUnaryClientInterceptorDoesNotRetryNonIdempotentMethod(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts: 3,
		IdempotentMethods: map[string]bool{
			"/trip.v1.Trip/SearchTrips": true,
		},
	}
	attempts := 0

	err := RetryUnaryClientInterceptor(policy)(context.Background(), "/order.v1.Order/CreateOrder", nil, nil, nil, func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		attempts++
		return status.Error(codes.Unavailable, "downstream unavailable")
	})

	require.Error(t, err)
	require.Equal(t, 1, attempts)
}

func TestTraceLoggingUnaryClientInterceptorWritesStandardFields(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core)
	ctx := context.WithValue(context.Background(), HeaderRequestID, "req-1")
	ctx = context.WithValue(ctx, HeaderTraceparent, "00-abc-def-01")
	ctx = context.WithValue(ctx, "user_id", int64(1001))

	err := TraceLoggingUnaryClientInterceptor(log)(ctx, "/trip.v1.Trip/SearchTrips", nil, nil, nil, func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		return status.Error(codes.Unavailable, "downstream unavailable")
	})

	require.Error(t, err)
	require.Len(t, logs.All(), 1)
	entry := logs.All()[0]
	require.Equal(t, "grpc client call", entry.Message)
	fields := entry.ContextMap()
	require.Equal(t, "/trip.v1.Trip/SearchTrips", fields["grpc_method"])
	require.Equal(t, "Unavailable", fields["grpc_code"])
	require.Equal(t, "req-1", fields["request_id"])
	require.Equal(t, "00-abc-def-01", fields["traceparent"])
	require.Equal(t, int64(1001), fields["user_id"])
	require.Contains(t, fields, "duration_ms")
}

func TestUnaryServerInterceptorExtractsMetadataMapsErrorAndLogs(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core)
	undo := zap.ReplaceGlobals(log)
	defer undo()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		HeaderUserID, "1001",
		HeaderRequestID, "req-1",
		HeaderTraceparent, "00-abc-def-01",
	))
	info := &grpc.UnaryServerInfo{FullMethod: "/order.v1.Order/GetOrderDetail"}

	_, err := UnaryServerInterceptor()(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		require.Equal(t, int64(1001), UserIDFromCtx(ctx))
		require.Equal(t, "req-1", RequestIDFromCtx(ctx))
		require.Equal(t, "00-abc-def-01", TraceparentFromCtx(ctx))
		return nil, apperrors.PermissionDenied("AUTHZ_RESOURCE_OWNER_DENIED", "forbidden")
	})

	require.Error(t, err)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.Len(t, logs.All(), 1)
	entry := logs.All()[0]
	require.Equal(t, "grpc server call", entry.Message)
	fields := entry.ContextMap()
	require.Equal(t, "/order.v1.Order/GetOrderDetail", fields["grpc_method"])
	require.Equal(t, "PermissionDenied", fields["grpc_code"])
	require.Equal(t, "req-1", fields["request_id"])
	require.Equal(t, "00-abc-def-01", fields["traceparent"])
	require.Equal(t, int64(1001), fields["user_id"])
	require.Contains(t, fields, "duration_ms")
}
