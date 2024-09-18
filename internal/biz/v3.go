package biz

import (
	"context"
	"github.com/BitofferHub/pkg/middlewares/snowflake"
)

func (uc *UrlMapUseCase) GenerateShortUrlV3(ctx context.Context, longUrl string) (string, error) {
	// 先查询缓存里面是否对应的短链
	shortUrl, err := uc.repo.GetShortUrlFormCache(ctx, longUrl)
	// 有错误 或者 缓存中有这个短链，直接返回
	if err != nil || shortUrl != "" {
		return shortUrl, err
	}

	// 再查询数据库里面是否有长链 对应的短链
	shortUrl, err = uc.repo.GetShortUrlFormDb(ctx, longUrl)
	// 有错误直接返回
	if err != nil {
		return "", err
	}

	// 如果有，顺便保存到缓存中
	if shortUrl != "" {
		uc.repo.SaveToCache(ctx, longUrl, shortUrl)
		return shortUrl, nil
	}

	// 还是没找到, 那就利用雪花算法生成ID
	id := snowflake.GenID()

	// 利用base62算法，生成短链
	shortUrl = generateShortUrl(id)

	// 将短链保存到布隆过滤器中
	err = uc.repo.SaveToBloomFilter(ctx, shortUrl)
	if err != nil {
		return "", err
	}

	// 保存到缓存中, 以便下次查询
	err = uc.repo.SaveToCache(ctx, longUrl, shortUrl)
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
func (uc *UrlMapUseCase) GetLongUrlV3(ctx context.Context, shortUrl string) (string, error) {
	// 从布隆过滤器中查询以及缓存中查询
	need, longUrl, err := uc.repo.FindShortUrlFormBloomFilterAndCache(ctx, shortUrl)
	// 发生错误 || longUrl不为空 || 不需要查询DB, 直接return
	if err != nil || longUrl != "" || need == 0 {
		return longUrl, err
	}

	// 从数据库中查询
	return uc.repo.GetLongUrlFormDb(ctx, shortUrl)
}
