package biz

import (
	"context"
	pb "github.com/BitofferHub/proto_center/api/shortUrlXsvr/v1"
	"strconv"
)

func (uc *UrlMapUseCase) GenerateShortUrlV2(ctx context.Context, longUrl string) (string, error) {
	// 先查询数据库里面是否有这个长链
	shortUrl, err := uc.repo.GetShortUrlFormDb(ctx, longUrl)
	if err != nil {
		return "", err
	}

	// 如果有，直接返回
	if shortUrl != "" {
		return shortUrl, nil
	}

	// 从缓存里面获取ID
	idStr, err := uc.repo.GenerateIdFromCache(ctx)

	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 利用base62算法，生成短链
	shortUrl = generateShortUrl(id)

	// 保存到布隆过滤器中
	err = uc.repo.SaveToBloomFilter(ctx, shortUrl)
	if err != nil {
		return "", err
	}

	// 保存到数据库
	_, err = uc.repo.CreateToDb(ctx, &UrlMap{ShortUrl: shortUrl, LongUrl: longUrl})
	if err != nil {
		return "", err
	}
	// 返回短链
	return shortUrl, nil
}

// GetLongUrl 获取长链
func (uc *UrlMapUseCase) GetLongUrlV2(ctx context.Context, shortUrl string) (string, error) {
	// 先从布隆过滤器中查询
	exist, err := uc.repo.FindShortUrlFormBloomFilter(ctx, shortUrl)
	if err != nil {
		return "", err
	}
	// 如果不存在于布隆过滤器中，那么也一定不存在于DB中
	if !exist {
		return "", pb.ErrorShortUrlNotFound("短链不存在")
	}

	return uc.repo.GetLongUrlFormDb(ctx, shortUrl)
}
