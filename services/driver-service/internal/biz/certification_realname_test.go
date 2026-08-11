package biz

import (
	"context"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ride-hailing/pkg/realname"
)

type stubRealNameVerifier struct {
	result realname.Result
	err    error
}

func (v stubRealNameVerifier) Verify(context.Context, realname.Request) (realname.Result, error) {
	return v.result, v.err
}

func TestSubmitCertificationVerifiesRealNameAndCreatesAudit(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.profiles[2001] = &DriverProfile{ID: 2001, CertificationStatus: CertificationStatusDraft}
	uc := NewDriverUsecase(node, zap.NewNop(), repo, stubRealNameVerifier{
		result: realname.Result{
			ErrorCode: 0,
			Matched:   true,
			City:      "杭州市",
			Area:      "浙江省杭州市西湖区",
		},
	})

	cert, err := uc.SubmitCertification(context.Background(), SubmitCertificationCommand{
		DriverID:    2001,
		RealName:    " 张三 ",
		IDCardNo:    " 330106199001011234 ",
		LicenseNo:   " DL001 ",
		LicenseType: " C1 ",
	})

	require.NoError(t, err)
	require.Equal(t, CertificationStatusPending, cert.Status)
	require.Equal(t, "330106199001011234", cert.IDCardNo)
	require.Equal(t, "C1", cert.LicenseType)
	require.Equal(t, "杭州市", cert.City)
	require.Equal(t, CertificationStatusPending, repo.profiles[2001].CertificationStatus)

	audit := repo.certificationAudits[2001]
	require.NotNil(t, audit)
	require.Equal(t, int64(2001), audit.UserID)
	require.Equal(t, "张三", audit.RealName)
	require.Equal(t, "330106199001011234", audit.CertNumber)
	require.Equal(t, "DL001", audit.DriverLicenseNo)
	require.Equal(t, "C1", audit.LicenseType)
	require.Equal(t, "杭州市", audit.City)
	require.Equal(t, 0, audit.Status)
	require.Equal(t, 1, audit.SubmitCount)
}

func TestSubmitCertificationRejectsRealNameMismatch(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	uc := NewDriverUsecase(node, zap.NewNop(), repo, stubRealNameVerifier{
		result: realname.Result{ErrorCode: 0, Matched: false, Reason: "Success"},
	})

	_, err = uc.SubmitCertification(context.Background(), SubmitCertificationCommand{
		DriverID:    2001,
		RealName:    "张三",
		IDCardNo:    "330106199001011234",
		LicenseNo:   "DL001",
		LicenseType: "C1",
	})

	require.ErrorIs(t, err, ErrRealNameNotMatched)
	require.Empty(t, repo.certifications)
}
