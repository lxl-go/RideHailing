package carpool

import "time"

type UserCouponView struct {
	ID             string    `json:"id"`
	CouponCode     string    `json:"couponCode"`
	CouponNo       string    `json:"couponNo"`
	UserID         string    `json:"userId"`
	UserType       string    `json:"userType"`
	Source         string    `json:"source"`
	Status         string    `json:"status"`
	OrderNo        string    `json:"orderNo"`
	OrderAmount    float64   `json:"orderAmount"`
	DiscountAmount float64   `json:"discountAmount"`
	Operator       string    `json:"operator"`
	IssuedAt       time.Time `json:"issuedAt"`
	UsedAt         time.Time `json:"usedAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CouponTemplate struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:coupon template id" json:"id"`
	CouponNo        string    `gorm:"column:coupon_no;type:varchar(64);not null;uniqueIndex:uk_marketing_coupon_no;comment:coupon number" json:"couponNo"`
	Name            string    `gorm:"column:name;type:varchar(128);not null;index:idx_marketing_coupon_keyword;comment:coupon name" json:"name"`
	CouponType      string    `gorm:"column:coupon_type;type:varchar(16);not null;index:idx_marketing_coupon_type_status;comment:cash discount free_ride" json:"couponType"`
	FaceValue       float64   `gorm:"column:face_value;type:decimal(10,2);not null;default:0;comment:cash amount" json:"faceValue"`
	DiscountRate    float64   `gorm:"column:discount_rate;type:decimal(4,2);not null;default:0;comment:discount rate" json:"discountRate"`
	ThresholdAmount float64   `gorm:"column:threshold_amount;type:decimal(10,2);not null;default:0;comment:min order amount" json:"thresholdAmount"`
	ValidFrom       time.Time `gorm:"column:valid_from;not null;index:idx_marketing_coupon_valid;comment:valid from" json:"validFrom"`
	ValidTo         time.Time `gorm:"column:valid_to;not null;index:idx_marketing_coupon_valid;comment:valid to" json:"validTo"`
	CityScope       string    `gorm:"column:city_scope;type:varchar(128);not null;default:'';index:idx_marketing_coupon_scope;comment:city scope" json:"cityScope"`
	ServiceScope    string    `gorm:"column:service_scope;type:varchar(32);not null;default:'';index:idx_marketing_coupon_scope;comment:carpool shuttle all" json:"serviceScope"`
	TimeScope       string    `gorm:"column:time_scope;type:varchar(128);not null;default:'';comment:time window" json:"timeScope"`
	Stackable       bool      `gorm:"column:stackable;type:tinyint;not null;default:0;comment:can stack" json:"stackable"`
	TotalStock      int       `gorm:"column:total_stock;type:int;not null;default:0;comment:total stock" json:"totalStock"`
	IssuedCount     int       `gorm:"column:issued_count;type:int;not null;default:0;comment:issued count" json:"issuedCount"`
	UsedCount       int       `gorm:"column:used_count;type:int;not null;default:0;comment:used count" json:"usedCount"`
	Status          string    `gorm:"column:status;type:varchar(16);not null;default:'draft';index:idx_marketing_coupon_type_status;comment:draft enabled disabled deleted" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (CouponTemplate) TableName() string {
	return "marketing_coupon_template"
}

type UserCoupon struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:user coupon id" json:"id"`
	CouponCode     string    `gorm:"column:coupon_code;type:varchar(64);not null;uniqueIndex:uk_marketing_user_coupon_code;comment:user coupon code" json:"couponCode"`
	CouponNo       string    `gorm:"column:coupon_no;type:varchar(64);not null;index:idx_marketing_user_coupon_tpl;comment:template number" json:"couponNo"`
	UserID         uint64    `gorm:"column:user_id;type:bigint;not null;index:idx_marketing_user_coupon_user;comment:user id" json:"userId"`
	UserType       string    `gorm:"column:user_type;type:varchar(16);not null;default:'passenger';comment:user type" json:"userType"`
	Source         string    `gorm:"column:source;type:varchar(16);not null;default:'manual';index:idx_marketing_user_coupon_source;comment:manual auto campaign referral" json:"source"`
	Status         string    `gorm:"column:status;type:varchar(16);not null;default:'unused';index:idx_marketing_user_coupon_status;comment:unused used expired refunded" json:"status"`
	OrderNo        string    `gorm:"column:order_no;type:varchar(64);not null;default:'';index:idx_marketing_user_coupon_order;comment:redeemed order no" json:"orderNo"`
	OrderAmount    float64   `gorm:"column:order_amount;type:decimal(10,2);not null;default:0;comment:order amount" json:"orderAmount"`
	DiscountAmount float64   `gorm:"column:discount_amount;type:decimal(10,2);not null;default:0;comment:discount amount" json:"discountAmount"`
	Operator       string    `gorm:"column:operator;type:varchar(64);not null;default:'';comment:operator" json:"operator"`
	IssuedAt       time.Time `gorm:"column:issued_at;not null;default:CURRENT_TIMESTAMP;comment:issued time" json:"issuedAt"`
	UsedAt         time.Time `gorm:"column:used_at;comment:used time" json:"usedAt"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (UserCoupon) TableName() string {
	return "marketing_user_coupon"
}

type MarketingCampaign struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:campaign id" json:"id"`
	CampaignNo string    `gorm:"column:campaign_no;type:varchar(64);not null;uniqueIndex:uk_marketing_campaign_no;comment:campaign no" json:"campaignNo"`
	Name       string    `gorm:"column:name;type:varchar(128);not null;index:idx_marketing_campaign_keyword;comment:campaign name" json:"name"`
	Channel    string    `gorm:"column:channel;type:varchar(32);not null;index:idx_marketing_campaign_channel;comment:social banner offline" json:"channel"`
	CouponNo   string    `gorm:"column:coupon_no;type:varchar(64);not null;index:idx_marketing_campaign_coupon;comment:coupon no" json:"couponNo"`
	StartAt    time.Time `gorm:"column:start_at;not null;index:idx_marketing_campaign_time;comment:start time" json:"startAt"`
	EndAt      time.Time `gorm:"column:end_at;not null;index:idx_marketing_campaign_time;comment:end time" json:"endAt"`
	Status     string    `gorm:"column:status;type:varchar(16);not null;default:'draft';index:idx_marketing_campaign_status;comment:draft running paused ended" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (MarketingCampaign) TableName() string {
	return "marketing_campaign"
}

type ReferralReward struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:referral reward id" json:"id"`
	ReferrerID   uint64    `gorm:"column:referrer_id;type:bigint;not null;index:idx_marketing_referral_referrer;comment:referrer id" json:"referrerId"`
	InviteeID    uint64    `gorm:"column:invitee_id;type:bigint;not null;index:idx_marketing_referral_invitee;comment:invitee id" json:"inviteeId"`
	CouponNo     string    `gorm:"column:coupon_no;type:varchar(64);not null;index:idx_marketing_referral_coupon;comment:coupon no" json:"couponNo"`
	RewardStatus string    `gorm:"column:reward_status;type:varchar(16);not null;default:'pending';index:idx_marketing_referral_status;comment:pending issued cancelled" json:"rewardStatus"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (ReferralReward) TableName() string {
	return "marketing_referral_reward"
}
