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

// NewUrlMapRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.NewUrlMapRepo
func NewUrlMapRepo(data *Data) biz.UrlMapRepo {
	return &urlMapRepo{
		data: data,
	}
}

func (r *urlMapRepo) GetLongUrlFormDb(ctx context.Context, shortUrl string) (string, error) {

	// 从数据库中获取
	urlMap := &biz.UrlMap{}
	err := r.data.db.WithContext(ctx).Select("long_url").Where("short_url = ?", shortUrl).First(urlMap).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	return urlMap.LongUrl, nil
}

func (r *urlMapRepo) GetShortUrlFormDb(ctx context.Context, longUrl string) (string, error) {

	urlMap := &biz.UrlMap{}
	// 从数据库中获取
	err := r.data.db.WithContext(ctx).Select("short_url").Where("long_url = ?", longUrl).First(urlMap).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	return urlMap.ShortUrl, nil

}

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

func (r *urlMapRepo) CreateToDb(ctx context.Context, x *biz.UrlMap) (int64, error) {
	db := r.data.db
	err := db.WithContext(ctx).Create(x).Error

	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return -1, err
	}
	// 主键ID也会被返回
	return x.ID, err
}

func (r *urlMapRepo) SaveToDb(ctx context.Context, x *biz.UrlMap) error {
	db := r.data.db
	// 更新到DB中
	err := db.WithContext(ctx).Model(biz.UrlMap{}).Where("long_url = ?", x.LongUrl).Update("short_url", x.ShortUrl).Error
	if err != nil {
		return err
	}
	return nil
}

// 从缓存里面获取ID
func (r *urlMapRepo) GenerateIdFromCache(ctx context.Context) (string, error) {
	id, err := r.data.cache.Incr(ctx, cacheIdKey)
	if err != nil {
		return "", err
	}
	return id, nil
}

// 保存到布隆过滤器
func (r *urlMapRepo) SaveToBloomFilter(ctx context.Context, shortUrl string) error {
	_, err := r.data.cache.BFAdd(ctx, shortUrlBloomFilterKey, shortUrl)
	if err != nil {
		return err
	}
	return nil
}

// 从布隆过滤器中查询短链是否存在
func (r *urlMapRepo) FindShortUrlFormBloomFilter(ctx context.Context, shortUrl string) (bool, error) {
	exists, err := r.data.cache.BFExists(ctx, shortUrlBloomFilterKey, shortUrl)
	if err != nil {
		return false, err
	}
	return exists, nil

}

// 保存短链到缓存中
func (r *urlMapRepo) SaveToCache(ctx context.Context, longUrl, shortUrl string) error {
	err := r.data.cache.Set(ctx, longUrlPrefix+longUrl, shortUrl, expireTime)
	if err != nil {
		return err
	}

	err = r.data.cache.Set(ctx, shortUrlPrefix+shortUrl, longUrl, expireTime)
	return err
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
