package biz

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"

	"ride-hailing/pkg/realname"
)

var vehiclePlatePattern = regexp.MustCompile(`^[\p{Han}][A-Z][A-Z0-9挂学警港澳]{5,6}$`)

const (
	DriverServiceStatusOffline  = 1
	DriverServiceStatusOnline   = 2
	DriverServiceStatusDisabled = 3
)

var vehiclePlateLegalPattern = regexp.MustCompile(`^[\p{Han}][A-Z][A-Z0-9挂学警港澳]{5,6}$`)

const (
	CertificationStatusDraft    = 1
	CertificationStatusPending  = 2
	CertificationStatusApproved = 3
	CertificationStatusRejected = 4
)

const (
	VehicleStatusActive   = 1
	VehicleStatusInactive = 2
)

const (
	VehicleAuditStatusPending       = 0
	VehicleAuditStatusApproved      = 1
	VehicleAuditStatusRejected      = 2
	VehicleAuditStatusDriverDeleted = 3
)

type DriverProfile struct {
	ID                  int64
	Name                string
	Phone               string
	AvatarURL           string
	ServiceStatus       int
	CertificationStatus int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type DriverCertification struct {
	ID               int64
	DriverID         int64
	RealName         string
	IDCardNo         string
	LicenseNo        string
	LicenseType      string
	City             string
	VehicleLicenseNo string
	VehiclePhotoURL  string
	FacePhotoURL     string
	Status           int
	RejectReason     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CertificationAudit struct {
	ID               int64
	UserID           int64
	UserRole         int
	CertType         int
	CertNumber       string
	RealName         string
	DriverLicenseNo  string
	LicenseType      string
	City             string
	FrontImageURL    string
	BackImageURL     string
	HandheldImageURL string
	Status           int
	SubmitCount      int
}

type DriverVehicle struct {
	ID           int64
	DriverID     int64
	AuditID      int64
	PlateNo      string
	Brand        string
	Model        string
	Color        string
	VehicleType  string
	Seats        int
	Status       int
	ReviewStatus int
	RejectReason string
	Source       string
	CanEdit      bool
	CanDelete    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DriverVehicleAudit struct {
	ID           int64
	DriverID     int64
	PlateNumber  string
	Brand        string
	Model        string
	Color        string
	Seats        int
	Status       int
	ReviewerID   int64
	RejectReason string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DriverLocationPoint struct {
	ID         int64
	DriverID   int64
	OrderID    int64
	Latitude   float64
	Longitude  float64
	Speed      float64
	Heading    float64
	ReportedAt time.Time
	CreatedAt  time.Time
}

type DriverMessage struct {
	ID        int64
	Topic     string
	Title     string
	Payload   string
	Delivered bool
	CreatedAt time.Time
}

type UpdateDriverCommand struct {
	ID            int64
	Name          string
	Phone         string
	AvatarURL     string
	ServiceStatus int
}

type SubmitCertificationCommand struct {
	DriverID         int64
	RealName         string
	IDCardNo         string
	LicenseNo        string
	LicenseType      string
	City             string
	VehicleLicenseNo string
	VehiclePhotoURL  string
	FacePhotoURL     string
}

type SaveVehicleCommand struct {
	DriverID    int64
	PlateNo     string
	Brand       string
	Model       string
	Color       string
	VehicleType string
	Seats       int
}

type UpdateVehicleCommand struct {
	DriverID    int64
	VehicleID   int64
	PlateNo     string
	Brand       string
	Model       string
	Color       string
	VehicleType string
	Seats       int
}

type DeleteVehicleCommand struct {
	DriverID  int64
	VehicleID int64
}

type ReportLocationCommand struct {
	DriverID   int64
	OrderID    int64
	Latitude   float64
	Longitude  float64
	Speed      float64
	Heading    float64
	ReportedAt time.Time
}

type TrackReplayQuery struct {
	DriverID int64
	OrderID  int64
	Page     int
	PageSize int
}

type TrackReplayResult struct {
	Total  int64
	Points []DriverLocationPoint
}

type ListMessagesQuery struct {
	DriverID int64
}

type AckMessageCommand struct {
	DriverID  int64
	MessageID int64
}

type DriverUsecase struct {
	node     *snowflake.Node
	log      *zap.Logger
	repo     DriverRepo
	verifier realname.Verifier
}

func NewDriverUsecase(node *snowflake.Node, log *zap.Logger, repo DriverRepo, verifiers ...realname.Verifier) *DriverUsecase {
	var verifier realname.Verifier
	if len(verifiers) > 0 {
		verifier = verifiers[0]
	}
	return &DriverUsecase{node: node, log: log, repo: repo, verifier: verifier}
}

func NewDriverUsecaseWithVerifier(node *snowflake.Node, log *zap.Logger, repo DriverRepo, verifier realname.Verifier) *DriverUsecase {
	return NewDriverUsecase(node, log, repo, verifier)
}

func (uc *DriverUsecase) EnsureDriver(ctx context.Context, id int64, phone string) (*DriverProfile, error) {
	if id <= 0 {
		return nil, ErrInvalidDriver
	}
	phone = strings.TrimSpace(phone)
	profile, err := uc.repo.GetProfileByID(ctx, id)
	if err == nil {
		if phone != "" && strings.TrimSpace(profile.Phone) == "" {
			profile.Phone = phone
			if err := uc.repo.UpdateProfile(ctx, profile); err != nil {
				uc.log.Error("backfill driver phone failed", zap.Error(err))
				return nil, err
			}
		}
		return profile, nil
	}
	if !errors.Is(err, ErrDriverNotFound) {
		return nil, err
	}
	profile = &DriverProfile{
		ID:                  id,
		Name:                fmt.Sprintf("Driver %d", id),
		Phone:               phone,
		ServiceStatus:       DriverServiceStatusOffline,
		CertificationStatus: CertificationStatusDraft,
	}
	if err := uc.repo.CreateProfile(ctx, profile); err != nil {
		uc.log.Error("create driver profile failed", zap.Error(err))
		return nil, err
	}
	return profile, nil
}

func (uc *DriverUsecase) GetDriver(ctx context.Context, id int64) (*DriverProfile, error) {
	if id <= 0 {
		return nil, ErrInvalidDriver
	}
	return uc.repo.GetProfileByID(ctx, id)
}

func (uc *DriverUsecase) UpdateDriver(ctx context.Context, cmd UpdateDriverCommand) (*DriverProfile, error) {
	if cmd.ID <= 0 {
		return nil, ErrInvalidDriver
	}
	profile, err := uc.repo.GetProfileByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	profile.Name = strings.TrimSpace(cmd.Name)
	profile.Phone = strings.TrimSpace(cmd.Phone)
	profile.AvatarURL = strings.TrimSpace(cmd.AvatarURL)
	if cmd.ServiceStatus != 0 {
		if !isValidDriverServiceStatus(cmd.ServiceStatus) {
			return nil, ErrInvalidDriver
		}
		profile.ServiceStatus = cmd.ServiceStatus
	}
	if profile.ServiceStatus == 0 {
		profile.ServiceStatus = DriverServiceStatusOffline
	}
	if profile.CertificationStatus == 0 {
		profile.CertificationStatus = CertificationStatusDraft
	}
	if err := uc.repo.UpdateProfile(ctx, profile); err != nil {
		uc.log.Error("update driver profile failed", zap.Error(err))
		return nil, err
	}
	return profile, nil
}

func isValidDriverServiceStatus(status int) bool {
	switch status {
	case DriverServiceStatusOffline, DriverServiceStatusOnline, DriverServiceStatusDisabled:
		return true
	default:
		return false
	}
}

func (uc *DriverUsecase) SubmitCertification(ctx context.Context, cmd SubmitCertificationCommand) (*DriverCertification, error) {
	realName := strings.TrimSpace(cmd.RealName)
	idCardNo := strings.TrimSpace(cmd.IDCardNo)
	licenseNo := strings.TrimSpace(cmd.LicenseNo)
	licenseType := strings.TrimSpace(cmd.LicenseType)
	if cmd.DriverID <= 0 || realName == "" || idCardNo == "" || licenseNo == "" || licenseType == "" {
		return nil, ErrInvalidDriver
	}
	if uc.verifier == nil {
		return nil, ErrRealNameUnavailable
	}
	verifyResult, err := uc.verifier.Verify(ctx, realname.Request{RealName: realName, IDCardNo: idCardNo})
	if err != nil {
		uc.log.Error("real-name authentication failed", zap.Error(err))
		return nil, ErrRealNameUnavailable
	}
	if !verifyResult.Matched {
		return nil, ErrRealNameNotMatched
	}
	city := strings.TrimSpace(cmd.City)
	if verifyResult.City != "" {
		city = verifyResult.City
	}
	cert := &DriverCertification{
		ID:               uc.node.Generate().Int64(),
		DriverID:         cmd.DriverID,
		RealName:         realName,
		IDCardNo:         idCardNo,
		LicenseNo:        licenseNo,
		LicenseType:      licenseType,
		City:             city,
		VehicleLicenseNo: strings.TrimSpace(cmd.VehicleLicenseNo),
		VehiclePhotoURL:  strings.TrimSpace(cmd.VehiclePhotoURL),
		FacePhotoURL:     strings.TrimSpace(cmd.FacePhotoURL),
		Status:           CertificationStatusPending,
	}
	if err := uc.repo.SaveCertification(ctx, cert); err != nil {
		uc.log.Error("save driver certification failed", zap.Error(err))
		return nil, err
	}
	if err := uc.repo.SaveCertificationAudit(ctx, &CertificationAudit{
		ID:               uc.node.Generate().Int64(),
		UserID:           cmd.DriverID,
		UserRole:         2,
		CertType:         1,
		CertNumber:       idCardNo,
		RealName:         realName,
		DriverLicenseNo:  licenseNo,
		LicenseType:      licenseType,
		City:             city,
		FrontImageURL:    strings.TrimSpace(cmd.VehiclePhotoURL),
		HandheldImageURL: strings.TrimSpace(cmd.FacePhotoURL),
		Status:           0,
		SubmitCount:      1,
	}); err != nil {
		uc.log.Error("save driver certification audit failed", zap.Error(err))
		return nil, err
	}
	profile, err := uc.EnsureDriver(ctx, cmd.DriverID, "")
	if err != nil {
		return nil, err
	}
	profile.CertificationStatus = CertificationStatusPending
	if err := uc.repo.UpdateProfile(ctx, profile); err != nil {
		uc.log.Error("update driver certification status failed", zap.Error(err))
		return nil, err
	}
	return cert, nil
}

func (uc *DriverUsecase) GetCertification(ctx context.Context, driverID int64) (*DriverCertification, error) {
	if driverID <= 0 {
		return nil, ErrInvalidDriver
	}
	return uc.repo.GetCertification(ctx, driverID)
}

func (uc *DriverUsecase) SaveVehicle(ctx context.Context, cmd SaveVehicleCommand) (*DriverVehicle, error) {
	plateNo, brand, model, color, vehicleType, seats, ok := normalizeVehicleInput(cmd.PlateNo, cmd.Brand, cmd.Model, cmd.Color, cmd.VehicleType, cmd.Seats)
	if cmd.DriverID <= 0 || !ok {
		return nil, ErrInvalidDriver
	}
	if model == "" {
		model = vehicleType
	}
	audit := &DriverVehicleAudit{
		ID:          uc.node.Generate().Int64(),
		DriverID:    cmd.DriverID,
		PlateNumber: plateNo,
		Brand:       brand,
		Model:       model,
		Color:       color,
		Seats:       seats,
		Status:      VehicleAuditStatusPending,
	}
	if err := uc.repo.SaveVehicleAudit(ctx, audit); err != nil {
		uc.log.Error("save driver vehicle audit failed", zap.Error(err))
		return nil, err
	}
	return vehicleFromAudit(audit), nil
}

func (uc *DriverUsecase) UpdateVehicle(ctx context.Context, cmd UpdateVehicleCommand) (*DriverVehicle, error) {
	plateNo, brand, model, color, vehicleType, seats, ok := normalizeVehicleInput(cmd.PlateNo, cmd.Brand, cmd.Model, cmd.Color, cmd.VehicleType, cmd.Seats)
	if cmd.DriverID <= 0 || cmd.VehicleID <= 0 || !ok {
		return nil, ErrInvalidDriver
	}
	vehicle, err := uc.repo.GetVehicleByID(ctx, cmd.VehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.DriverID != cmd.DriverID || vehicle.Status != VehicleStatusActive {
		return nil, ErrVehicleNotFound
	}
	if model == "" {
		model = vehicleType
	}
	audit := &DriverVehicleAudit{
		ID:          vehicle.ID,
		DriverID:    cmd.DriverID,
		PlateNumber: plateNo,
		Brand:       brand,
		Model:       model,
		Color:       color,
		Seats:       seats,
		Status:      VehicleAuditStatusPending,
	}
	if err := uc.repo.SaveVehicleAudit(ctx, audit); err != nil {
		uc.log.Error("save driver vehicle update audit failed", zap.Error(err))
		return nil, err
	}
	return vehicleFromAudit(audit), nil
}

func normalizeVehicleInput(plateNo, brand, model, color, vehicleType string, seats int) (string, string, string, string, string, int, bool) {
	plateNo = strings.ToUpper(strings.TrimSpace(plateNo))
	brand = strings.TrimSpace(brand)
	model = strings.TrimSpace(model)
	color = strings.TrimSpace(color)
	vehicleType = strings.TrimSpace(vehicleType)
	if !isValidPlateNo(plateNo) || !isValidVehicleText(brand, 64, true) || !isValidVehicleText(model, 64, false) || !isValidVehicleText(color, 32, false) || !isValidVehicleText(vehicleType, 32, false) {
		return "", "", "", "", "", 0, false
	}
	if seats < 1 || seats > 9 {
		return "", "", "", "", "", 0, false
	}
	return plateNo, brand, model, color, vehicleType, seats, true
}

func isValidPlateNo(plateNo string) bool {
	return vehiclePlateLegalPattern.MatchString(plateNo)
}

func isValidVehicleText(value string, maxLen int, required bool) bool {
	if required && value == "" {
		return false
	}
	if len([]rune(value)) > maxLen {
		return false
	}
	for _, r := range value {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

func (uc *DriverUsecase) DeleteVehicle(ctx context.Context, cmd DeleteVehicleCommand) error {
	if cmd.DriverID <= 0 || cmd.VehicleID <= 0 {
		return ErrInvalidDriver
	}
	vehicle, err := uc.repo.GetVehicleByID(ctx, cmd.VehicleID)
	if err != nil {
		return err
	}
	if vehicle.DriverID != cmd.DriverID || vehicle.Status != VehicleStatusActive {
		return ErrVehicleNotFound
	}
	vehicle.Status = VehicleStatusInactive
	if err := uc.repo.UpdateVehicle(ctx, vehicle); err != nil {
		uc.log.Error("delete driver vehicle failed", zap.Error(err))
		return err
	}
	if err := uc.repo.MarkVehicleAuditDriverDeleted(ctx, cmd.DriverID, vehicle.PlateNo); err != nil {
		uc.log.Error("mark driver vehicle audit deleted failed", zap.Error(err))
		return err
	}
	return nil
}

func (uc *DriverUsecase) ListVehicles(ctx context.Context, driverID int64) ([]DriverVehicle, error) {
	if driverID <= 0 {
		return nil, ErrInvalidDriver
	}
	vehicles, err := uc.repo.ListVehicles(ctx, driverID)
	if err != nil {
		return nil, err
	}
	for i := range vehicles {
		applyFormalVehicleMeta(&vehicles[i])
	}
	audits, err := uc.repo.ListVehicleAudits(ctx, driverID)
	if err != nil {
		return nil, err
	}
	vehicles = hideFormalVehiclesCoveredByOpenAudits(vehicles, audits)
	for _, audit := range audits {
		if audit.Status == VehicleAuditStatusPending {
			vehicles = append(vehicles, *vehicleFromAudit(&audit))
		}
	}
	return vehicles, nil
}

func hideFormalVehiclesCoveredByOpenAudits(vehicles []DriverVehicle, audits []DriverVehicleAudit) []DriverVehicle {
	if len(vehicles) == 0 || len(audits) == 0 {
		return vehicles
	}
	blockedIDs := make(map[int64]struct{})
	blockedPlates := make(map[string]struct{})
	for _, audit := range audits {
		if audit.Status != VehicleAuditStatusPending && audit.Status != VehicleAuditStatusRejected {
			continue
		}
		if audit.ID > 0 {
			blockedIDs[audit.ID] = struct{}{}
		}
		plate := strings.ToUpper(strings.TrimSpace(audit.PlateNumber))
		if plate != "" {
			blockedPlates[plate] = struct{}{}
		}
	}
	if len(blockedIDs) == 0 && len(blockedPlates) == 0 {
		return vehicles
	}
	filtered := vehicles[:0]
	for _, vehicle := range vehicles {
		if _, ok := blockedIDs[vehicle.ID]; ok {
			continue
		}
		plate := strings.ToUpper(strings.TrimSpace(vehicle.PlateNo))
		if _, ok := blockedPlates[plate]; ok {
			continue
		}
		filtered = append(filtered, vehicle)
	}
	return filtered
}

func (uc *DriverUsecase) ListMessages(ctx context.Context, query ListMessagesQuery) ([]DriverMessage, error) {
	if query.DriverID <= 0 {
		return nil, ErrInvalidDriver
	}
	return uc.repo.ListDriverMessages(ctx, query.DriverID)
}

func (uc *DriverUsecase) AckMessage(ctx context.Context, cmd AckMessageCommand) error {
	if cmd.DriverID <= 0 || cmd.MessageID <= 0 {
		return ErrInvalidDriver
	}
	return uc.repo.AckDriverMessage(ctx, cmd.DriverID, cmd.MessageID)
}

func applyFormalVehicleMeta(vehicle *DriverVehicle) {
	if vehicle == nil {
		return
	}
	vehicle.Source = "vehicle"
	vehicle.AuditID = 0
	vehicle.ReviewStatus = VehicleAuditStatusApproved
	vehicle.CanEdit = vehicle.Status == VehicleStatusActive
	vehicle.CanDelete = vehicle.Status == VehicleStatusActive
}

func vehicleFromAudit(audit *DriverVehicleAudit) *DriverVehicle {
	if audit == nil {
		return nil
	}
	seats := audit.Seats
	if seats < 1 || seats > 9 {
		seats = 4
	}
	return &DriverVehicle{
		ID:           audit.ID,
		DriverID:     audit.DriverID,
		AuditID:      audit.ID,
		PlateNo:      audit.PlateNumber,
		Brand:        audit.Brand,
		Model:        audit.Model,
		VehicleType:  audit.Model,
		Color:        audit.Color,
		Seats:        seats,
		Status:       audit.Status,
		ReviewStatus: audit.Status,
		RejectReason: audit.RejectReason,
		Source:       "audit",
		CanEdit:      false,
		CanDelete:    false,
		CreatedAt:    audit.CreatedAt,
		UpdatedAt:    audit.UpdatedAt,
	}
}

func (uc *DriverUsecase) ReportLocation(ctx context.Context, cmd ReportLocationCommand) (*DriverLocationPoint, error) {
	if cmd.DriverID <= 0 || cmd.OrderID <= 0 || cmd.Latitude < -90 || cmd.Latitude > 90 || cmd.Longitude < -180 || cmd.Longitude > 180 {
		return nil, ErrInvalidDriverLocation
	}
	reportedAt := cmd.ReportedAt
	if reportedAt.IsZero() {
		reportedAt = time.Now()
	}
	point := &DriverLocationPoint{
		ID:         uc.node.Generate().Int64(),
		DriverID:   cmd.DriverID,
		OrderID:    cmd.OrderID,
		Latitude:   cmd.Latitude,
		Longitude:  cmd.Longitude,
		Speed:      cmd.Speed,
		Heading:    cmd.Heading,
		ReportedAt: reportedAt,
	}
	if err := uc.repo.SaveLocation(ctx, point); err != nil {
		uc.log.Error("save driver location failed", zap.Error(err))
		return nil, err
	}
	return point, nil
}

func (uc *DriverUsecase) ReplayTrack(ctx context.Context, query TrackReplayQuery) (*TrackReplayResult, error) {
	if query.DriverID <= 0 {
		return nil, ErrInvalidDriver
	}
	if query.OrderID <= 0 {
		return nil, ErrInvalidDriverLocation
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 500 {
		query.PageSize = 200
	}
	points, total, err := uc.repo.ReplayTrack(ctx, query)
	if err != nil {
		uc.log.Error("replay driver track failed", zap.Error(err))
		return nil, err
	}
	return &TrackReplayResult{Total: total, Points: points}, nil
}
