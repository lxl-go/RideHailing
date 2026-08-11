package data

import (
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ride-hailing/services/trip-service/internal/biz"
	"ride-hailing/services/trip-service/internal/conf"
)

type Data struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewData(db *gorm.DB, client *redis.Client) *Data {
	return &Data{DB: db, Redis: client}
}

func NewRedis(c *conf.Redis) (*redis.Client, error) {
	if c == nil || c.Addr == "" {
		return nil, fmt.Errorf("redis config is required")
	}
	dialTimeout, _ := time.ParseDuration(c.DialTimeout)
	readTimeout, _ := time.ParseDuration(c.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(c.WriteTimeout)
	client := redis.NewClient(&redis.Options{Addr: c.Addr, Username: c.Username, Password: c.Password, DB: c.DB, PoolSize: c.PoolSize, DialTimeout: dialTimeout, ReadTimeout: readTimeout, WriteTimeout: writeTimeout})
	return client, nil
}

func NewDB(c *conf.Data, logger *zap.Logger) (*gorm.DB, error) {
	if c == nil || c.Database == nil {
		return nil, fmt.Errorf("database config is required")
	}
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&tripModel{}, &couponTemplateModel{}, &userCouponModel{}, &tripDemandModel{}); err != nil {
		return nil, err
	}
	return db, nil
}

func NewSnowflakeNode(c *conf.Server) (*snowflake.Node, error) {
	if c == nil {
		return snowflake.NewNode(1)
	}
	node := c.SnowflakeNode
	if node == 0 {
		node = 1
	}
	return snowflake.NewNode(node)
}

var ProviderSet = wire.NewSet(
	NewDB,
	NewRedis,
	NewData,
	NewSnowflakeNode,
	NewAMapClient,
	NewTripRepo,
	wire.Bind(new(biz.TripRepo), new(*TripRepo)),
)
