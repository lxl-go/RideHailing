package biz

import (
	"context"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCompleteOrderMovesDeliveringOrderToCompleted(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusDelivering,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	err = uc.CompleteOrder(context.Background(), OrderActionCommand{
		OrderID:        5001,
		ActorID:        2001,
		IdempotencyKey: "complete-5001",
	})

	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, repo.orders[0].Status)
}

func TestCompleteOrderRejectsWrongDriver(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusDelivering,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	err = uc.CompleteOrder(context.Background(), OrderActionCommand{
		OrderID:        5001,
		ActorID:        2999,
		IdempotencyKey: "complete-wrong-driver",
	})

	require.ErrorIs(t, err, ErrNotTripOwner)
	require.Equal(t, OrderStatusDelivering, repo.orders[0].Status)
}
