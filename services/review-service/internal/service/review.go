package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	reviewv1 "ride-hailing/services/review-service/api/review/v1"
	"ride-hailing/services/review-service/internal/biz"
)

type ReviewService struct {
	reviewv1.UnimplementedReviewServiceServer
	uc *biz.ReviewUsecase
}

func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{uc: uc}
}

func (s *ReviewService) SubmitReview(ctx context.Context, req *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error) {
	review, err := s.uc.SubmitReview(ctx, biz.SubmitReviewCommand{
		OrderID:    req.OrderId,
		FromUserID: req.FromUserId,
		ToUserID:   req.ToUserId,
		Rating:     int(req.Rating),
		Content:    req.Content,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &reviewv1.SubmitReviewReply{ReviewId: review.ID}, nil
}

func (s *ReviewService) GetMyReview(ctx context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error) {
	review, hasReview, err := s.uc.GetMyReview(ctx, req.OrderId, req.FromUserId)
	if err != nil {
		return nil, mapError(err)
	}
	return &reviewv1.GetMyReviewReply{HasReview: hasReview, Review: reviewToProto(review)}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidReview), errors.Is(err, biz.ErrOrderNotCompleted):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrReviewAlreadyExists), errors.Is(err, biz.ErrReviewerNotAuthorized):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func reviewToProto(review *biz.Review) *reviewv1.ReviewItem {
	if review == nil {
		return nil
	}
	return &reviewv1.ReviewItem{
		Id:         review.ID,
		OrderId:    review.OrderID,
		FromUserId: review.FromUserID,
		ToUserId:   review.ToUserID,
		Rating:     int32(review.Rating),
		Content:    review.Content,
		CreatedAt:  review.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
