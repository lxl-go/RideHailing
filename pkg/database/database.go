package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Init 带超时的 DB 初始化（对标文档 1.2 cmd 启动规范：所有初始化加 context.WithTimeout）
func Init(ctx context.Context, fn func() (*gorm.DB, error), timeout time.Duration) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan *gorm.DB, 1)
	errCh := make(chan error, 1)

	go func() {
		db, err := fn()
		if err != nil {
			errCh <- err
			return
		}
		done <- db
	}()

	select {
	case db := <-done:
		return db, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("database init timeout after %v: %w", timeout, ctx.Err())
	}
}

// Paginate 通用分页（对标文档 4.5 节分页规范）
func Paginate(page, pageSize int32) (offset, limit int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return int((page - 1) * pageSize), int(pageSize)
}

// WithContext 安全地设置 DB 上下文（对标文档 ctx 透传规范）
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx != nil {
		return db.WithContext(ctx)
	}
	return db
}
