package data

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ride-hailing/services/review-service/internal/biz"
	"ride-hailing/services/review-service/internal/conf"
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
	if err := db.AutoMigrate(&reviewModel{}); err != nil {
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

var ProviderSet = wire.NewSet(
	NewDB,
	NewData,
	NewSnowflakeNode,
	NewReviewRepo,
	wire.Bind(new(biz.ReviewRepo), new(*ReviewRepo)),
)
