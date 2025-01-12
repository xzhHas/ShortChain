package data

import (
	"context"
	"errors"

	"github.com/BitofferHub/shortUrlX/internal/biz"
	"gorm.io/gorm"
)

type urlMapRepo struct {
	data *Data
}

func NewUrlMapRepo(data *Data) biz.UrlMapRepo {
	return &urlMapRepo{
		data: data,
	}
}

// MySQL操作相关

// GetLongUrlFormDb 通过short_url从数据库中获取长链接
func (r *urlMapRepo) GetLongUrlFormDb(ctx context.Context, shortUrl string) (string, error) {
	urlMap := &biz.UrlMap{}
	err := r.data.db.WithContext(ctx).Select("long_url").Where("short_url = ?", shortUrl).First(urlMap).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	return urlMap.LongUrl, nil
}

// GetShortUrlFormDb 通过long_url从数据库中获取短链
func (r *urlMapRepo) GetShortUrlFormDb(ctx context.Context, longUrl string) (string, error) {
	urlMap := &biz.UrlMap{}
	// SELECT short_url FROM url_map WHERE long_url = ?
	err := r.data.db.WithContext(ctx).Select("short_url").Where("long_url = ?", longUrl).First(urlMap).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	return urlMap.ShortUrl, nil
}

// CreateToDb 创建到DB中
func (r *urlMapRepo) CreateToDb(ctx context.Context, x *biz.UrlMap) (int64, error) {
	err := r.data.db.WithContext(ctx).Create(x).Error
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return -1, err
	}
	// 主键ID也会被返回
	return x.ID, err
}

// SaveToDb 更新short_url到MySQL中
func (r *urlMapRepo) SaveToDb(ctx context.Context, x *biz.UrlMap) error {
	// UPDATE url_map SET short_url = ? WHERE long_url = ?
	err := r.data.db.WithContext(ctx).Model(biz.UrlMap{}).Where("long_url = ?", x.LongUrl).Update("short_url", x.ShortUrl).Error
	if err != nil {
		return err
	}
	return nil
}

// Redis 操作相关

// GetShortUrlFormCache 从缓存中获取短链short_url
func (r *urlMapRepo) GetShortUrlFormCache(ctx context.Context, longUrl string) (string, error) {
	// 1. 先从缓存中获取
	res, exists, err := r.data.cache.Get(ctx, longUrlPrefix+longUrl)
	if err != nil {
		return "", err
	}
	// 如果存在, 直接返回
	if exists {
		return res, nil
	}
	return "", nil
}

// GenerateIdFromCache 通过cacheIdKey从缓存中获取ID
func (r *urlMapRepo) GenerateIdFromCache(ctx context.Context) (string, error) {
	// redis的Incr命令，用于生成自增ID
	// 将cacheIdKey对应的value每次自增1，返回这个值做ID
	id, err := r.data.cache.Incr(ctx, cacheIdKey)
	if err != nil {
		return "", err
	}
	return id, nil
}

// SaveToCache 将长链接和短链接保存到缓存中
func (r *urlMapRepo) SaveToCache(ctx context.Context, longUrl, shortUrl string) error {
	// 这里存储了两个k-v 键值对
	// longUrlPrefix+longUrl -> shortUrl
	// shortUrlPrefix+shortUrl -> longUrl
	err := r.data.cache.Set(ctx, longUrlPrefix+longUrl, shortUrl, expireTime)
	if err != nil {
		return err
	}
	return r.data.cache.Set(ctx, shortUrlPrefix+shortUrl, longUrl, expireTime)
}

// 布隆过滤器

// SaveToBloomFilter （布隆过滤器）将短链保存到布隆过滤器
func (r *urlMapRepo) SaveToBloomFilter(ctx context.Context, shortUrl string) error {
	// shortUrlBloomFilterKey 是这个布隆过滤器的名字
	// shortUrl是存储的值
	_, err := r.data.cache.BFAdd(ctx, shortUrlBloomFilterKey, shortUrl)
	if err != nil {
		return err
	}
	return nil
}

// FindShortUrlFormBloomFilter 从布隆过滤器中查询短链是否存在
func (r *urlMapRepo) FindShortUrlFormBloomFilter(ctx context.Context, shortUrl string) (bool, error) {
	exists, err := r.data.cache.BFExists(ctx, shortUrlBloomFilterKey, shortUrl)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *urlMapRepo) FindShortUrlFormBloomFilterAndCache(ctx context.Context, shortUrl string) (int64, string, error) {
	// 利用Lua脚本，先从布隆过滤器中查询，如果存在，再从缓存中查询
	res, err := r.data.cache.EvalResults(ctx, findShortUrlFormBloomFilterAndCacheLua, []string{shortUrlBloomFilterKey, shortUrlPrefix + shortUrl}, shortUrl)
	if err != nil {
		return 0, "", err
	}

	resSLice := res.([]interface{})
	need := resSLice[0].(int64)
	longUrl := resSLice[1].(string)
	return need, longUrl, nil
}
