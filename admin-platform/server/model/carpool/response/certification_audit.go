package response

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"ride-hailing/admin-server/model/carpool"
)

type CertificationAuditResponse struct {
	ID                  string         `json:"id"`
	UserID              string         `json:"userId"`
	UserRole            int            `json:"userRole"`
	CertType            int            `json:"certType"`
	CertNumber          string         `json:"certNumber"`
	RealName            string         `json:"realName"`
	DriverLicenseNo     string         `json:"driverLicenseNo"`
	LicenseType         string         `json:"licenseType"`
	City                string         `json:"city"`
	FrontImageURL       string         `json:"frontImageUrl"`
	BackImageURL        string         `json:"backImageUrl"`
	HandheldImageURL    string         `json:"handheldImageUrl"`
	Status              int            `json:"status"`
	ReviewerID          string         `json:"reviewerId"`
	RejectReason        string         `json:"rejectReason"`
	SubmitCount         int            `json:"submitCount"`
	ReviewedAt          *time.Time     `json:"reviewedAt"`
	ReviewDurationHours int            `json:"reviewDurationHours"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `json:"-"`
}

func NewCertificationAuditResponse(audit carpool.CertificationAudit) CertificationAuditResponse {
	return CertificationAuditResponse{
		ID:                  strconv.FormatInt(audit.ID, 10),
		UserID:              strconv.FormatInt(audit.UserID, 10),
		UserRole:            audit.UserRole,
		CertType:            audit.CertType,
		CertNumber:          audit.CertNumber,
		RealName:            audit.RealName,
		DriverLicenseNo:     audit.DriverLicenseNo,
		LicenseType:         audit.LicenseType,
		City:                audit.City,
		FrontImageURL:       audit.FrontImageURL,
		BackImageURL:        audit.BackImageURL,
		HandheldImageURL:    audit.HandheldImageURL,
		Status:              audit.Status,
		ReviewerID:          strconv.FormatInt(audit.ReviewerID, 10),
		RejectReason:        audit.RejectReason,
		SubmitCount:         audit.SubmitCount,
		ReviewedAt:          audit.ReviewedAt,
		ReviewDurationHours: audit.ReviewDurationHours,
		CreatedAt:           audit.CreatedAt,
		UpdatedAt:           audit.UpdatedAt,
		DeletedAt:           audit.DeletedAt,
	}
}

func NewCertificationAuditResponses(items []carpool.CertificationAudit) []CertificationAuditResponse {
	responses := make([]CertificationAuditResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewCertificationAuditResponse(item))
	}
	return responses
}
