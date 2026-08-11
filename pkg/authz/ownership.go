package authz

import (
	"errors"
	"fmt"

	apperrors "ride-hailing/pkg/errors"
)

var ErrForbiddenResource = errors.New("forbidden resource")

func RequireOwner(actorID, ownerID int64, resource string) error {
	if actorID <= 0 {
		return apperrors.Unauthenticated("AUTHZ_ACTOR_REQUIRED", "actor is required")
	}
	if ownerID <= 0 || actorID != ownerID {
		return apperrors.Wrap(
			apperrors.KindPermissionDenied,
			"AUTHZ_RESOURCE_OWNER_DENIED",
			fmt.Sprintf("not allowed to access %s", resourceName(resource)),
			ErrForbiddenResource,
		)
	}
	return nil
}

func RequireDataScope(allowedIDs []int64, targetID int64, resource string) error {
	if targetID <= 0 {
		return apperrors.InvalidArgument("AUTHZ_TARGET_REQUIRED", "target resource is required")
	}
	for _, id := range allowedIDs {
		if id == targetID {
			return nil
		}
	}
	return apperrors.Wrap(
		apperrors.KindPermissionDenied,
		"AUTHZ_DATA_SCOPE_DENIED",
		fmt.Sprintf("not allowed to access %s", resourceName(resource)),
		ErrForbiddenResource,
	)
}

func resourceName(resource string) string {
	if resource == "" {
		return "resource"
	}
	return resource
}
