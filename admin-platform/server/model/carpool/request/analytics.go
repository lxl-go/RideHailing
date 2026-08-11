package request

type AnalyticsSearch struct {
	Period      string `json:"period" form:"period"`
	StartDate   string `json:"startDate" form:"startDate"`
	EndDate     string `json:"endDate" form:"endDate"`
	ServiceType string `json:"serviceType" form:"serviceType"`
}
