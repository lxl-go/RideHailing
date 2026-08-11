package carpool

import (
	"context"
	"strconv"

	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/model/carpool"
	"ride-hailing/admin-server/utils/logger"
)

type AuditNotifier struct{}

func (n *AuditNotifier) NotifyCertificationAuditResult(ctx context.Context, audit *carpool.CertificationAudit, result string) {
	if audit == nil {
		return
	}

	if len(global.GVA_CONFIG.Kafka.Brokers) == 0 || global.GVA_CONFIG.Kafka.Topic == "" || global.GVA_KAFKA == nil {
		logCertificationNotification(ctx, audit.ID, "kafka disabled, skip certification audit result message")
	} else {
		if global.GVA_LOG != nil {
			logger.WithCtx(ctx).Mod("carpool").Field("audit_id", audit.ID).Field("topic", global.GVA_CONFIG.Kafka.Topic).Info("reserved kafka certification audit notification")
		}
	}

	if !global.GVA_CONFIG.SMS.Enabled || global.GVA_SMS == nil {
		logCertificationNotification(ctx, audit.ID, "sms disabled, skip certification audit sms")
		return
	}

	if err := global.GVA_SMS.SendSMS("", map[string]string{
		"auditId": strconv.FormatInt(audit.ID, 10),
		"result":  result,
	}); err != nil {
		logger.WithCtx(ctx).Mod("carpool").Err(err).Field("audit_id", audit.ID).Warn("certification audit sms send failed")
	}
}

func logCertificationNotification(ctx context.Context, auditID int64, message string) {
	if global.GVA_LOG == nil {
		return
	}
	logger.WithCtx(ctx).Mod("carpool").Field("audit_id", auditID).Info(message)
}
