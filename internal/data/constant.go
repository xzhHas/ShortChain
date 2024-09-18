package data

import "time"

const (
	projectKey = "shortUrlX-"

	shortUrlPrefix = projectKey + "ShortUrl:"
	longUrlPrefix  = projectKey + "LongUrl:"

	// 过期时间
	expireTime = time.Hour * 24
	cacheIdKey = projectKey + "IncrId"

	shortUrlBloomFilterKey = projectKey + "BloomFilter-ShortUrl"

	// 访问布隆过滤器以及缓存
	findShortUrlFormBloomFilterAndCacheLua = `
  local bloomKey = KEYS[1]
  local cacheKey = KEYS[2]
  local bloomVal = ARGV[1]
-- 检查val是否存在于布隆过滤器对应的bloomKey中
  local exists = redis.call('BF.EXISTS', bloomKey, bloomVal)

-- 如果bloomVal不存在于布隆过滤器中，直接返回空字符串, 返回0代表不需要查db了
  if exists == 0 then
      return {0, ''}
  end

-- 如果bloomVal存在于布隆过滤器中，查询cacheKey
  local value = redis.call('GET', cacheKey)

-- 如果cacheKey存在，返回对应的值，否则返回空字符串
  if value then
      return {0, value}
  else
      return {1, ''}
  end
`
)
