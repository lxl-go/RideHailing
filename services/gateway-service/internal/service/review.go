package service

import (
	"context"

	"ride-hailing/services/gateway-service/internal/biz"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
)

type ReviewService struct {
	uc *biz.ReviewUsecase
}

func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{uc: uc}
}

func (s *ReviewService) SubmitReview(ctx context.Context, req *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error) {
	return s.uc.SubmitReview(ctx, req)
}

func (s *ReviewService) GetMyReview(ctx context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error) {
	return s.uc.GetMyReview(ctx, req)
}
