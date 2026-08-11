package biz

import (
	"context"

	"ride-hailing/services/gateway-service/internal/data"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
)

type ReviewUsecase struct {
	client data.ReviewClient
}

func NewReviewUsecase(client data.ReviewClient) *ReviewUsecase {
	return &ReviewUsecase{client: client}
}

func (uc *ReviewUsecase) SubmitReview(ctx context.Context, req *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error) {
	return uc.client.SubmitReview(ctx, req)
}

func (uc *ReviewUsecase) GetMyReview(ctx context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error) {
	return uc.client.GetMyReview(ctx, req)
}
