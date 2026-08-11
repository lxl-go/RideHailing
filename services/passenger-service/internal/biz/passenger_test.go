package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakePassengerRepo struct {
	items map[int64]*PassengerProfile
}

func newFakePassengerRepo() *fakePassengerRepo {
	return &fakePassengerRepo{items: map[int64]*PassengerProfile{}}
}

func (r *fakePassengerRepo) GetByID(_ context.Context, id int64) (*PassengerProfile, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, ErrPassengerNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakePassengerRepo) Create(_ context.Context, profile *PassengerProfile) error {
	copy := *profile
	r.items[profile.ID] = &copy
	return nil
}

func (r *fakePassengerRepo) Update(_ context.Context, profile *PassengerProfile) error {
	copy := *profile
	r.items[profile.ID] = &copy
	return nil
}

func TestEnsurePassengerCreatesDefaultProfile(t *testing.T) {
	repo := newFakePassengerRepo()
	uc := NewPassengerUsecase(zap.NewNop(), repo)

	profile, err := uc.EnsurePassenger(context.Background(), 1001, "13800138000")

	require.NoError(t, err)
	require.Equal(t, int64(1001), profile.ID)
	require.Equal(t, "Passenger 1001", profile.Nickname)
	require.Equal(t, "13800138000", profile.Phone)
	require.Equal(t, PassengerStatusEnabled, profile.Status)
	require.Len(t, repo.items, 1)
}

func TestEnsurePassengerBackfillsPhoneWhenMissing(t *testing.T) {
	repo := newFakePassengerRepo()
	repo.items[1001] = &PassengerProfile{ID: 1001, Nickname: "Alice", Phone: "", Status: PassengerStatusEnabled}
	uc := NewPassengerUsecase(zap.NewNop(), repo)

	profile, err := uc.EnsurePassenger(context.Background(), 1001, "13800138000")

	require.NoError(t, err)
	require.Equal(t, "13800138000", profile.Phone)
	require.Equal(t, "Alice", profile.Nickname)
}

func TestEnsurePassengerKeepsExistingPhone(t *testing.T) {
	repo := newFakePassengerRepo()
	repo.items[1001] = &PassengerProfile{ID: 1001, Nickname: "Alice", Phone: "13900000000", Status: PassengerStatusEnabled}
	uc := NewPassengerUsecase(zap.NewNop(), repo)

	profile, err := uc.EnsurePassenger(context.Background(), 1001, "13800138000")

	require.NoError(t, err)
	require.Equal(t, "13900000000", profile.Phone)
}

func TestUpdatePassengerTrimsProfileFields(t *testing.T) {
	repo := newFakePassengerRepo()
	repo.items[1001] = &PassengerProfile{ID: 1001, Status: PassengerStatusEnabled}
	uc := NewPassengerUsecase(zap.NewNop(), repo)

	profile, err := uc.UpdatePassenger(context.Background(), UpdatePassengerCommand{
		ID:                1001,
		Nickname:          "  Alice  ",
		Phone:             "  13800138000  ",
		CommonAddress:     "  Shanghai  ",
		PaymentPreference: "  wechat  ",
	})

	require.NoError(t, err)
	require.Equal(t, "Alice", profile.Nickname)
	require.Equal(t, "13800138000", profile.Phone)
	require.Equal(t, "Shanghai", profile.CommonAddress)
	require.Equal(t, "wechat", profile.PaymentPreference)
}
