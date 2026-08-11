package initialize

import (
	"testing"

	"ride-hailing/admin-server/config"
)

func TestMysqlDialectorConfigDisablesDatetimePrecision(t *testing.T) {
	cfg := mysqlDialectorConfig(config.Mysql{})

	if !cfg.DisableDatetimePrecision {
		t.Fatal("mysql auto migration should disable datetime precision so CURRENT_TIMESTAMP defaults work on MySQL 8")
	}
}
