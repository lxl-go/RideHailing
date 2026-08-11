package carpool

import "time"

type FinanceTransaction struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:交易流水ID" json:"id"`
	OrderNo        string    `gorm:"column:order_no;type:varchar(64);not null;uniqueIndex:uk_finance_order;comment:订单号" json:"orderNo"`
	DriverID       uint64    `gorm:"column:driver_id;type:bigint;not null;index:idx_finance_driver;comment:司机ID" json:"driverId"`
	PassengerID    uint64    `gorm:"column:passenger_id;type:bigint;not null;index:idx_finance_passenger;comment:乘客ID" json:"passengerId"`
	Amount         float64   `gorm:"column:amount;type:decimal(10,2);not null;default:0;comment:交易金额" json:"amount"`
	PaymentMethod  string    `gorm:"column:payment_method;type:varchar(32);not null;default:'wallet';comment:支付方式" json:"paymentMethod"`
	Status         string    `gorm:"column:status;type:varchar(32);not null;default:'success';index:idx_finance_status;comment:交易状态" json:"status"`
	Abnormal       bool      `gorm:"column:abnormal;type:tinyint;not null;default:0;index:idx_finance_abnormal;comment:是否异常" json:"abnormal"`
	AbnormalReason string    `gorm:"column:abnormal_reason;type:varchar(255);not null;default:'';comment:异常原因" json:"abnormalReason"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

func (FinanceTransaction) TableName() string {
	return "finance_transaction"
}

type FinanceRefund struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:退款记录ID" json:"id"`
	OrderNo      string    `gorm:"column:order_no;type:varchar(64);not null;index:idx_finance_refund_order;comment:订单号" json:"orderNo"`
	RefundNo     string    `gorm:"column:refund_no;type:varchar(64);not null;uniqueIndex:uk_finance_refund;comment:退款单号" json:"refundNo"`
	RefundAmount float64   `gorm:"column:refund_amount;type:decimal(10,2);not null;default:0;comment:退款金额" json:"refundAmount"`
	Status       string    `gorm:"column:status;type:varchar(16);not null;default:'processing';index:idx_finance_refund_status;comment:退款状态" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

func (FinanceRefund) TableName() string {
	return "finance_refund"
}
