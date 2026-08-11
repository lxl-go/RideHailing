package service

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/driver-service/internal/biz"
)

type DriverService struct {
	driverv1.UnimplementedDriverServiceServer
	uc *biz.DriverUsecase
}

func NewDriverService(uc *biz.DriverUsecase) *DriverService {
	return &DriverService{uc: uc}
}

func (s *DriverService) EnsureDriver(ctx context.Context, req *driverv1.EnsureDriverRequest) (*driverv1.DriverProfileReply, error) {
	driver, err := s.uc.EnsureDriver(ctx, req.Id, req.Phone)
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.DriverProfileReply{Driver: driverToProto(driver)}, nil
}

func (s *DriverService) GetDriver(ctx context.Context, req *driverv1.GetDriverRequest) (*driverv1.DriverProfileReply, error) {
	driver, err := s.uc.GetDriver(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.DriverProfileReply{Driver: driverToProto(driver)}, nil
}

func (s *DriverService) UpdateDriver(ctx context.Context, req *driverv1.UpdateDriverRequest) (*driverv1.DriverProfileReply, error) {
	driver, err := s.uc.UpdateDriver(ctx, biz.UpdateDriverCommand{
		ID:            req.Id,
		Name:          req.Name,
		Phone:         req.Phone,
		AvatarURL:     req.AvatarUrl,
		ServiceStatus: int(req.ServiceStatus),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.DriverProfileReply{Driver: driverToProto(driver)}, nil
}

func (s *DriverService) SubmitCertification(ctx context.Context, req *driverv1.SubmitCertificationRequest) (*driverv1.CertificationReply, error) {
	cert, err := s.uc.SubmitCertification(ctx, biz.SubmitCertificationCommand{
		DriverID:         req.Id,
		RealName:         req.RealName,
		IDCardNo:         req.IdCardNo,
		LicenseNo:        req.LicenseNo,
		LicenseType:      req.LicenseType,
		City:             req.City,
		VehicleLicenseNo: req.VehicleLicenseNo,
		VehiclePhotoURL:  req.VehiclePhotoUrl,
		FacePhotoURL:     req.FacePhotoUrl,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.CertificationReply{Certification: certToProto(cert)}, nil
}

func (s *DriverService) GetCertification(ctx context.Context, req *driverv1.GetCertificationRequest) (*driverv1.CertificationReply, error) {
	cert, err := s.uc.GetCertification(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.CertificationReply{Certification: certToProto(cert)}, nil
}

func (s *DriverService) SaveVehicle(ctx context.Context, req *driverv1.SaveVehicleRequest) (*driverv1.VehicleReply, error) {
	vehicle, err := s.uc.SaveVehicle(ctx, biz.SaveVehicleCommand{
		DriverID:    req.Id,
		PlateNo:     req.PlateNo,
		Brand:       req.Brand,
		Model:       req.Model,
		Color:       req.Color,
		VehicleType: req.VehicleType,
		Seats:       int(req.Seats),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.VehicleReply{Vehicle: vehicleToProto(vehicle)}, nil
}

func (s *DriverService) UpdateVehicle(ctx context.Context, req *driverv1.UpdateVehicleRequest) (*driverv1.VehicleReply, error) {
	vehicle, err := s.uc.UpdateVehicle(ctx, biz.UpdateVehicleCommand{
		DriverID:    req.DriverId,
		VehicleID:   req.VehicleId,
		PlateNo:     req.PlateNo,
		Brand:       req.Brand,
		Model:       req.Model,
		Color:       req.Color,
		VehicleType: req.VehicleType,
		Seats:       int(req.Seats),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.VehicleReply{Vehicle: vehicleToProto(vehicle)}, nil
}

func (s *DriverService) DeleteVehicle(ctx context.Context, req *driverv1.DeleteVehicleRequest) (*driverv1.DeleteVehicleReply, error) {
	if err := s.uc.DeleteVehicle(ctx, biz.DeleteVehicleCommand{DriverID: req.DriverId, VehicleID: req.VehicleId}); err != nil {
		return nil, mapError(err)
	}
	return &driverv1.DeleteVehicleReply{Success: true}, nil
}

func (s *DriverService) ListVehicles(ctx context.Context, req *driverv1.ListVehiclesRequest) (*driverv1.ListVehiclesReply, error) {
	items, err := s.uc.ListVehicles(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*driverv1.DriverVehicle, len(items))
	for i := range items {
		out[i] = vehicleToProto(&items[i])
	}
	return &driverv1.ListVehiclesReply{Items: out}, nil
}

func (s *DriverService) ListMessages(ctx context.Context, req *driverv1.ListMessagesRequest) (*driverv1.ListMessagesReply, error) {
	items, err := s.uc.ListMessages(ctx, biz.ListMessagesQuery{DriverID: req.DriverId})
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*driverv1.DriverMessage, len(items))
	for i := range items {
		out[i] = messageToProto(&items[i])
	}
	return &driverv1.ListMessagesReply{Items: out}, nil
}

func (s *DriverService) AckMessage(ctx context.Context, req *driverv1.AckMessageRequest) (*driverv1.AckMessageReply, error) {
	if err := s.uc.AckMessage(ctx, biz.AckMessageCommand{DriverID: req.DriverId, MessageID: req.MessageId}); err != nil {
		return nil, mapError(err)
	}
	return &driverv1.AckMessageReply{Success: true}, nil
}

func (s *DriverService) ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error) {
	reportedAt, err := parseOptionalTime(req.ReportedAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reported_at")
	}
	point, err := s.uc.ReportLocation(ctx, biz.ReportLocationCommand{
		DriverID:   req.DriverId,
		OrderID:    req.OrderId,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Speed:      req.Speed,
		Heading:    req.Heading,
		ReportedAt: reportedAt,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &driverv1.DriverLocationReply{Location: locationToProto(point)}, nil
}

func (s *DriverService) ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error) {
	result, err := s.uc.ReplayTrack(ctx, biz.TrackReplayQuery{
		DriverID: req.DriverId,
		OrderID:  req.OrderId,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]*driverv1.DriverLocationPoint, len(result.Points))
	for i := range result.Points {
		items[i] = locationToProto(&result.Points[i])
	}
	return &driverv1.ReplayDriverTrackReply{Total: result.Total, Items: items}, nil
}

func mapError(err error) error {
	return mapErrorChineseMessage(err)
}

func mapErrorChinese(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidDriver), errors.Is(err, biz.ErrInvalidDriverLocation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrRealNameNotMatched):
		return status.Error(codes.PermissionDenied, "实名认证未通过：真实姓名和身份证号不匹配，或身份证信息库无记录，请核对后重试")
	case errors.Is(err, biz.ErrRealNameUnavailable):
		return status.Error(codes.Unavailable, "实名认证服务暂不可用，请稍后重试")
	case errors.Is(err, biz.ErrCertificationNotFound):
		return status.Error(codes.NotFound, "暂未提交司机认证资料")
	case errors.Is(err, biz.ErrVehiclePlateInUse):
		return status.Error(codes.AlreadyExists, "该车牌已被其他司机提交或绑定")
	case errors.Is(err, biz.ErrDriverNotFound), errors.Is(err, biz.ErrVehicleNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapErrorChineseMessage(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidDriver), errors.Is(err, biz.ErrInvalidDriverLocation):
		return status.Error(codes.InvalidArgument, "提交信息不合法，请检查后重新提交")
	case errors.Is(err, biz.ErrRealNameNotMatched):
		return status.Error(codes.PermissionDenied, "实名认证未通过：真实姓名和身份证号不匹配，或身份证信息库无记录，请核对后重试")
	case errors.Is(err, biz.ErrRealNameUnavailable):
		return status.Error(codes.Unavailable, "实名认证服务暂不可用，请稍后重试")
	case errors.Is(err, biz.ErrCertificationNotFound):
		return status.Error(codes.NotFound, "暂未提交司机认证资料")
	case errors.Is(err, biz.ErrVehiclePlateInUse):
		return status.Error(codes.AlreadyExists, "该车牌已被其他司机提交或绑定")
	case errors.Is(err, biz.ErrDriverNotFound), errors.Is(err, biz.ErrVehicleNotFound):
		return status.Error(codes.NotFound, "车辆不存在或已被处理，请刷新后重试")
	default:
		return status.Error(codes.Internal, "系统繁忙，请稍后重试")
	}
}

func driverToProto(driver *biz.DriverProfile) *driverv1.DriverProfile {
	if driver == nil {
		return nil
	}
	return &driverv1.DriverProfile{
		Id:                  driver.ID,
		Name:                driver.Name,
		Phone:               driver.Phone,
		AvatarUrl:           driver.AvatarURL,
		ServiceStatus:       int32(driver.ServiceStatus),
		CertificationStatus: int32(driver.CertificationStatus),
		CreatedAt:           formatTime(driver.CreatedAt),
		UpdatedAt:           formatTime(driver.UpdatedAt),
	}
}

func certToProto(cert *biz.DriverCertification) *driverv1.DriverCertification {
	if cert == nil {
		return nil
	}
	return &driverv1.DriverCertification{
		Id:               cert.ID,
		DriverId:         cert.DriverID,
		RealName:         cert.RealName,
		IdCardNo:         cert.IDCardNo,
		LicenseNo:        cert.LicenseNo,
		LicenseType:      cert.LicenseType,
		City:             cert.City,
		VehicleLicenseNo: cert.VehicleLicenseNo,
		VehiclePhotoUrl:  cert.VehiclePhotoURL,
		FacePhotoUrl:     cert.FacePhotoURL,
		Status:           int32(cert.Status),
		RejectReason:     cert.RejectReason,
		CreatedAt:        formatTime(cert.CreatedAt),
		UpdatedAt:        formatTime(cert.UpdatedAt),
	}
}

func vehicleToProto(vehicle *biz.DriverVehicle) *driverv1.DriverVehicle {
	if vehicle == nil {
		return nil
	}
	return &driverv1.DriverVehicle{
		Id:           vehicle.ID,
		DriverId:     vehicle.DriverID,
		PlateNo:      vehicle.PlateNo,
		Brand:        vehicle.Brand,
		Model:        vehicle.Model,
		Color:        vehicle.Color,
		VehicleType:  vehicle.VehicleType,
		Seats:        int32(vehicle.Seats),
		Status:       int32(vehicle.Status),
		CreatedAt:    formatTime(vehicle.CreatedAt),
		UpdatedAt:    formatTime(vehicle.UpdatedAt),
		AuditId:      vehicle.AuditID,
		ReviewStatus: int32(vehicle.ReviewStatus),
		RejectReason: vehicle.RejectReason,
		Source:       vehicle.Source,
		CanEdit:      vehicle.CanEdit,
		CanDelete:    vehicle.CanDelete,
	}
}

func messageToProto(message *biz.DriverMessage) *driverv1.DriverMessage {
	if message == nil {
		return nil
	}
	return &driverv1.DriverMessage{
		Id:        message.ID,
		Topic:     message.Topic,
		Title:     message.Title,
		Payload:   message.Payload,
		Delivered: message.Delivered,
		CreatedAt: formatTime(message.CreatedAt),
	}
}

func locationToProto(point *biz.DriverLocationPoint) *driverv1.DriverLocationPoint {
	if point == nil {
		return nil
	}
	return &driverv1.DriverLocationPoint{
		Id:         point.ID,
		DriverId:   point.DriverID,
		OrderId:    point.OrderID,
		Latitude:   point.Latitude,
		Longitude:  point.Longitude,
		Speed:      point.Speed,
		Heading:    point.Heading,
		ReportedAt: formatTime(point.ReportedAt),
		CreatedAt:  formatTime(point.CreatedAt),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func parseOptionalTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}
