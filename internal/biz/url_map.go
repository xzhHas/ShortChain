package biz

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type UrlMap struct {
	ID        int64 `gorm:"primaryKey"`
	LongUrl   string
	ShortUrl  string
	CreatedAt time.Time
}

// TableName 表名
func (p *UrlMap) TableName() string {
	return "url_map"
}

func (p *UrlMap) BeforeCreate(tx *gorm.DB) error {
	p.CreatedAt = time.Now()
	return nil
}

// UrlMapRepo is a Greater repo.
type UrlMapRepo interface {
	// 查询长链（从DB）
	GetLongUrlFormDb(context.Context, string) (string, error)

	// 查询短链（从DB）
	GetShortUrlFormDb(context.Context, string) (string, error)
	// 查询短链（从缓存）
	GetShortUrlFormCache(context.Context, string) (string, error)

	CreateToDb(context.Context, *UrlMap) (int64, error)

	SaveToDb(context.Context, *UrlMap) error
	SaveToCache(context.Context, string, string) error

	// 从缓存中获取ID
	GenerateIdFromCache(context.Context) (string, error)
	// 保存短链到布隆过滤器中
	SaveToBloomFilter(context.Context, string) error

	FindShortUrlFormBloomFilter(context.Context, string) (bool, error)

	FindShortUrlFormBloomFilterAndCache(context.Context, string) (int64, string, error)
}

type UrlMapUseCase struct {
	repo UrlMapRepo
	tm   Transaction
}

func NewUrlMapUseCase(repo UrlMapRepo, tm Transaction) *UrlMapUseCase {
	return &UrlMapUseCase{repo: repo, tm: tm}
}
