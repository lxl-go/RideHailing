package global

type SMSClient interface {
	SendSMS(phone string, templateParams map[string]string) error
}
