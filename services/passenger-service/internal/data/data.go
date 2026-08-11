package data

import (
	"fmt"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ride-hailing/services/passenger-service/internal/biz"
	"ride-hailing/services/passenger-service/internal/conf"
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
	if err := db.AutoMigrate(&passengerModel{}); err != nil {
		return nil, err
	}
	return db, nil
}

var ProviderSet = wire.NewSet(
	NewDB,
	NewData,
	NewPassengerRepo,
	wire.Bind(new(biz.PassengerRepo), new(*PassengerRepo)),
)
