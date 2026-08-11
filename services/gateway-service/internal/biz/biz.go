package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewAuthUsecase, NewTripUsecase, NewOrderUsecase, NewReviewUsecase, NewPassengerUsecase, NewDriverUsecase)
