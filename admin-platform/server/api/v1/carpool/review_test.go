package carpool

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestAuditOperationFailureMessageForRecordNotFound(t *testing.T) {
	msg := auditOperationFailureMessage(gorm.ErrRecordNotFound)

	if msg != "审核记录不存在或已被处理，请刷新列表后重试" {
		t.Fatalf("msg = %q", msg)
	}
}

func TestAuditOperationFailureMessageForOtherErrors(t *testing.T) {
	msg := auditOperationFailureMessage(errors.New("db timeout"))

	if msg != "操作失败，请稍后重试" {
		t.Fatalf("msg = %q", msg)
	}
}
