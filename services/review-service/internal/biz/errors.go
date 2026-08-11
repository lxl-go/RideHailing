package biz

import "errors"

var (
	ErrInvalidReview         = errors.New("invalid review")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderNotCompleted     = errors.New("order is not completed")
	ErrReviewNotFound        = errors.New("review not found")
	ErrReviewAlreadyExists   = errors.New("review already exists for this order")
	ErrReviewerNotAuthorized = errors.New("not authorized to review this order")
)
