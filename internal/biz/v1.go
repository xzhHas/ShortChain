package biz

import "context"

// GenerateShortUrlV1
func (uc *UrlMapUseCase) GenerateShortUrlV1(ctx context.Context, longUrl string) (string, error) {
	// 1. 先查询数据库里面是否有这个长链
	shortUrl, err := uc.repo.GetShortUrlFormDb(ctx, longUrl)
	if err != nil {
		return "", err
	}

	// 如果有，直接返回
	if shortUrl != "" {
		return shortUrl, nil
	}

	// 2. 如果没有，在数据库里面创建一条记录
	id, err := uc.repo.CreateToDb(ctx, &UrlMap{LongUrl: longUrl})
	if err != nil {
		return "", err
	}

	// 3. 利用base62算法 生成短链
	shortUrl = generateShortUrl(id)

	// 4. 更新对应记录，将短链存储到DB
	err = uc.repo.SaveToDb(ctx, &UrlMap{ShortUrl: shortUrl, LongUrl: longUrl})
	if err != nil {
		return "", err
	}

	// 5. 返回短链
	return shortUrl, nil
}

// GetLongUrl 获取长链
func (uc *UrlMapUseCase) GetLongUrlV1(ctx context.Context, shortUrl string) (string, error) {
	return uc.repo.GetLongUrlFormDb(ctx, shortUrl)
}
