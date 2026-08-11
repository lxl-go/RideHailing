package carpool

import "time"

type Review struct {
	ID         int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:主键ID"`
	OrderID    int64     `json:"orderId" gorm:"column:order_id;type:bigint;not null;uniqueIndex:uk_order_from_user;index:idx_order;comment:订单ID"`
	FromUserID int64     `json:"fromUserId" gorm:"column:from_user_id;type:bigint;not null;uniqueIndex:uk_order_from_user;index:idx_from_user;comment:评价人ID"`
	ToUserID   int64     `json:"toUserId" gorm:"column:to_user_id;type:bigint;not null;index:idx_to_user;comment:被评价人ID"`
	Rating     int       `json:"rating" gorm:"column:rating;type:tinyint;not null;comment:评分 1-5"`
	Content    string    `json:"content" gorm:"column:content;type:text;comment:评价内容"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
}

func (Review) TableName() string {
	return "carpool_review"
}
