package data

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/passenger-service/internal/biz"
)

func TestPassengerRepoCreateGetAndUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&passengerModel{}))

	repo := NewPassengerRepo(db, zap.NewNop())
	err = repo.Create(context.Background(), &biz.PassengerProfile{
		ID:       1001,
		Nickname: "Passenger 1001",
		Status:   biz.PassengerStatusEnabled,
	})
	require.NoError(t, err)

	profile, err := repo.GetByID(context.Background(), 1001)
	require.NoError(t, err)
	require.Equal(t, "Passenger 1001", profile.Nickname)

	profile.Nickname = "Alice"
	profile.Phone = "13800138000"
	require.NoError(t, repo.Update(context.Background(), profile))

	updated, err := repo.GetByID(context.Background(), 1001)
	require.NoError(t, err)
	require.Equal(t, "Alice", updated.Nickname)
	require.Equal(t, "13800138000", updated.Phone)
}
