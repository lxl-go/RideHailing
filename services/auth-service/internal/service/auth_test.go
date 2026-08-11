package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ride-hailing/services/auth-service/internal/biz"
)

func TestMapErrorReturnsResourceExhaustedForSMSRateLimits(t *testing.T) {
	for _, err := range []error{biz.ErrSMSCodeTooFrequent, biz.ErrSMSLoginLocked} {
		mapped := mapError(err)

		require.Equal(t, codes.ResourceExhausted, status.Code(mapped))
		require.Equal(t, err.Error(), status.Convert(mapped).Message())
	}
}

func TestMapErrorReturnsChineseMessageForSMSSendFailure(t *testing.T) {
	mapped := mapError(errors.New("ihuyi sms rejected: code=401 msg=invalid account"))

	require.Equal(t, codes.Internal, status.Code(mapped))
	require.Equal(t, "验证码发送失败，请检查短信服务配置或稍后重试", status.Convert(mapped).Message())
}
