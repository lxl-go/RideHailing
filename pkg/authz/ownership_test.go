package authz

import (
	"errors"
	"testing"

	apperrors "ride-hailing/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestRequireOwnerAllowsMatchingOwner(t *testing.T) {
	err := RequireOwner(1001, 1001, "order")
	require.NoError(t, err)
}

func TestRequireOwnerRejectsDifferentOwner(t *testing.T) {
	err := RequireOwner(1001, 2002, "order")

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrForbiddenResource))

	var appErr *apperrors.Error
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, apperrors.KindPermissionDenied, appErr.Kind)
	require.Equal(t, "AUTHZ_RESOURCE_OWNER_DENIED", appErr.Code)
}

func TestRequireDataScopeAllowsExplicitTarget(t *testing.T) {
	err := RequireDataScope([]int64{10, 20, 30}, 20, "driver")
	require.NoError(t, err)
}

func TestRequireDataScopeRejectsOutOfScopeTarget(t *testing.T) {
	err := RequireDataScope([]int64{10, 20, 30}, 40, "driver")

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrForbiddenResource))
}

func TestRequireValidActorRejectsEmptyActor(t *testing.T) {
	err := RequireOwner(0, 1001, "order")

	require.Error(t, err)
	var appErr *apperrors.Error
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, apperrors.KindUnauthenticated, appErr.Kind)
	require.Equal(t, "AUTHZ_ACTOR_REQUIRED", appErr.Code)
}
