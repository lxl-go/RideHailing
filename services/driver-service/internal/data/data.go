package data

import (
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ride-hailing/pkg/realname"
	"ride-hailing/services/driver-service/internal/biz"
	"ride-hailing/services/driver-service/internal/conf"
)

type Data struct {
	DB *gorm.DB
}

func NewData(db *gorm.DB) *Data {
	return &Data{DB: db}
}

func NewDB(c *conf.Data, logger *zap.Logger) (*gorm.DB, error) {
	if c == nil || c.Database == nil {
		return nil, fmt.Errorf("database config is required")
	}
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&driverProfileModel{}, &driverCertificationModel{}, &certificationAuditModel{}, &driverVehicleModel{}, &vehicleAuditModel{}, &driverLocationPointModel{}); err != nil {
		return nil, err
	}
	return db, nil
}

func NewSnowflakeNode(c *conf.Server) (*snowflake.Node, error) {
	if c == nil || c.SnowflakeNode == 0 {
		return snowflake.NewNode(1)
	}
	return snowflake.NewNode(c.SnowflakeNode)
}

func NewRealNameVerifier(c *conf.RealName) realname.Verifier {
	if c == nil || c.Tencent == nil || !c.Tencent.Enabled {
		return nil
	}
	timeout, _ := time.ParseDuration(c.Tencent.Timeout)
	return realname.NewTencentClient(realname.TencentConfig{
		SecretID:  c.Tencent.SecretID,
		SecretKey: c.Tencent.SecretKey,
		Endpoint:  c.Tencent.Endpoint,
		Timeout:   timeout,
	})
}

var ProviderSet = wire.NewSet(
	NewDB,
	NewData,
	NewSnowflakeNode,
	NewRealNameVerifier,
	NewDriverRepo,
	wire.Bind(new(biz.DriverRepo), new(*DriverRepo)),
)
