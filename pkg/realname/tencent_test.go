package realname

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTencentClientTreatsIsOKAsBusinessDecision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NotEmpty(t, r.Header.Get("Authorization"))
		require.NotEmpty(t, r.Header.Get("request-id"))
		require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		require.NoError(t, r.ParseForm())
		require.Equal(t, "110101199001011111", r.PostForm.Get("cardNo"))
		require.Equal(t, "张三", r.PostForm.Get("realName"))

		_, _ = w.Write([]byte(`{
			"error_code": 0,
			"reason": "Success",
			"result": {
				"realname": "张*",
				"idcard": "1101**********1111",
				"isok": false,
				"IdCardInfor": {
					"province": "北京市",
					"city": "北京市",
					"district": "东城区",
					"area": "北京市东城区",
					"sex": "男",
					"birthday": "1990-1-1"
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewTencentClient(TencentConfig{
		SecretID:  "secret-id",
		SecretKey: "secret-key",
		Endpoint:  server.URL,
	})

	result, err := client.Verify(context.Background(), Request{
		RealName: "张三",
		IDCardNo: "110101199001011111",
	})

	require.NoError(t, err)
	require.False(t, result.Matched)
	require.Equal(t, 0, result.ErrorCode)
	require.Equal(t, "北京市", result.City)
	require.Equal(t, "北京市东城区", result.Area)
}

func TestTencentClientReturnsBusinessCodeForNoExist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error_code":206501,"reason":"NoExistERROR","result":{"isok":false,"IdCardInfor":null}}`))
	}))
	defer server.Close()

	client := NewTencentClient(TencentConfig{
		SecretID:  "secret-id",
		SecretKey: "secret-key",
		Endpoint:  server.URL,
	})

	result, err := client.Verify(context.Background(), Request{
		RealName: "不存在",
		IDCardNo: "110101199001019999",
	})

	require.NoError(t, err)
	require.False(t, result.Matched)
	require.Equal(t, 206501, result.ErrorCode)
	require.Equal(t, "NoExistERROR", result.Reason)
}
