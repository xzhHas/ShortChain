# Shorthain 短链系统开发文档

## 1 认识短链

什么是短链？

为什么需要短链？

短链原理？

## 2 如何生成短链

### 2.1 哈希算法

go：crypto [https://pkg.go.dev/crypto](https://pkg.go.dev/crypto)

什么是哈希算法？

### 2.2 唯一ID算法

#### 2.2.1 MySQL

主键自增ID。适合单体应用，并发性能差。

#### 2.2.2 Redis

INCR自增原子操作。但是需要考虑持久化。

#### 2.2.3 雪花算法

![alt text](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455993.png)
**雪花算法的实现原理：**

雪花算法是一种随时间变化的分布式全局唯一ID算法，其生成的ID可以看做是一个64位的正整数，除了最高位，将剩余的63位分别分为41位的时间戳，10位的机器ID以及12位的自增序列号。

我们不采用MySQL的主键自增ID和redsi的incr的自增ID，而是使用本地雪花算法的形式直接生成ID，这样性能更高。

## 3 状态码和接口的设计

### 3.1 状态码选择

https://tsejx.github.io/javascript-guidebook/computer-networks/http/http-status-code/

**301 Moved Permanently：**表示永久重定向，说明请求的资源已经不存在了，需改用新的 URL 再次访问。

**302 Found：**表示临时重定向，说明请求的资源还在，但暂时需要用另一个 URL 来访问。

**304 Not** **Modified**：（未修改）自从上次请求后，请求的网页未修改过。 服务器返回此响应时，不会返回网页内容。

**为什么选择302？**



### 3.2 接口设计

| Method | path             | Request  | Response     |
| ------ | ---------------- | -------- | ------------ |
| GET    | /api/{short_url} | null     | 302,long_url |
| POST   | /api/shorten     | long_url | short_url    |

## 4 存储设计

### 4.1 数据库设计

```sql
DROP TABLE IF EXISTS `url_map`;
CREATE TABLE `url_map`  (
`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,
`long_url` varchar(250) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '长链',
`short_url` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '短链',
`created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
PRIMARY KEY (`id`) USING BTREE,
UNIQUE INDEX `idx_long_url_short_url`(`long_url` ASC, `short_url` ASC) USING BTREE,
INDEX `idx_short_url_long_url`(`short_url` ASC, `long_url` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;
```

注：联合索引最左匹配原则

1. MySQL会从联合索引最左边的索引开始匹配查询条件，从左到右匹配，如果查询条件没有使用到某个列，那么该列右边的列全部失效。
2. 当查询条件使用了某个列，但是该列的值包含范围查询，范围查询的字段可以用到联合索引，但是范围查询字段后面的字段无法用到联合索引。

### 4.2 缓存设计

![image-20250104121734908](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455544.png)

![image-20250104205228043](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455515.png)

## 5 代码实现

在线演示功能：[https://github.com/xzhHas/ShortChain/tree/main](https://github.com/xzhHas/ShortChain/tree/main)

### 5.1 MySQL版本 v1

直接依赖数据库的自增主键，将其base62即可。相对比较简单。

![alt text](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455752.png)
```go
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

// GetLongUrlV1 获取长链
func (uc *UrlMapUseCase) GetLongUrlV1(ctx context.Context, shortUrl string) (string, error) {
	return uc.repo.GetLongUrlFormDb(ctx, shortUrl)
}
```

### 5.2 MySQL+Redis v2

相比v1版本的get和post都是直接打到数据库而言，我们在v2中做了以下优化，

1. 使用INCR获取自增ID，再使用base62转换；
2. 使用Redis的布隆过滤器；

**流程**

![alt text](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455013.png)
**什么是布隆过滤器？**

布隆过滤器是用来检索一个元素是否存在于一个集合中，是由很长的二进制向量与一系列随机函数构成。

![alt text](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455511.png)
[布隆过滤器](https://javaguide.cn/cs-basics/data-structure/bloom-filter.html#%E4%BB%80%E4%B9%88%E6%98%AF%E5%B8%83%E9%9A%86%E8%BF%87%E6%BB%A4%E5%99%A8)

**关于数据库和Redis的配置**

```yaml
data:
  database:
    addr: 
    user: 
    password: 
    database: 
    max_idle_conn: 100
    max_open_conn: 150
    max_idle_time: 30
    slow_threshold_millisecond: 10 # SQL执行超过10ms，就算慢sql

  redis:
    addr: 
    password: 
    db: 9
    pool_size: 20
    bloom_filter_size: 10000 # 布隆过滤器大小
    error_rate: 0.01 # 布隆过滤器错误率
    read_timeout: 2s
    write_timeout: 2s
```

**核心代码**

```go
// GenerateShortUrlV2 生成短链
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



// GetLongUrlV2 获取长链
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
```

### 5.3 MySQL+Redis+雪花算法 v3

1. 不采用服务器生成ID的方式，使用本地生成ID，效率更高。
2. Redis做缓存提高性能。

**POST**

![image-20250104153047540](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455842.png)

**GET**

![image-20250104153103092](https://golang-code.oss-cn-beijing.aliyuncs.com/images/202501071455191.png)

### 5.4 一些算法

```go
// base62Encode 将十进制数字转换为62进制
func base62Encode(decimal int64) string {
    var result strings.Builder
    for decimal > 0 {
        remainder := decimal % 62
        result.WriteByte(characters[remainder])
        decimal /= 62
    }
    return result.String()
}

// shuffleString 打乱字符顺序，避免被猜测到
func shuffleString(input string) string {
    // 洗牌算法，打乱字符串的字符顺序
    chars := strings.Split(input, "")
    for i := len(chars) - 1; i > 0; i-- {
        // 获取一个0-i之间的随机索引
        j := globalRand.Intn(i + 1)
        chars[i], chars[j] = chars[j], chars[i]
    }
    return strings.Join(chars, "")
}

// 雪花算法
// 雪花算法设置初始时间，机器id
func Init(startTime time.Time, machineID int64) {
	startTimeStr := startTime.Format("2006-01-02 00:00:00")
	var st time.Time
	// 时间为UTC时间，比中国慢8个小时
	st, err := time.Parse("2006-01-02 00:00:00", startTimeStr)
	if err != nil {
		panic(err)
	}
	snowflake.Epoch = st.UnixNano() / 1000000 // 毫秒
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		panic(err)
	}
	return
}

func (n *Node) Generate() ID {
    // 加锁以确保线程安全
    n.mu.Lock()

    // 计算当前时间戳（单位：毫秒）
    // 从定义的 epoch (纪元时间) 开始，计算到现在的时间差，并将其转换为毫秒
    now := time.Since(n.epoch).Nanoseconds() / 1000000

    // 如果当前时间戳与上次生成 ID 的时间戳相同
    if now == n.time {
        // 增加步进序列（step）
        // 使用步进掩码 (stepMask) 进行按位与操作，确保步进序列不超过允许的最大值
        n.step = (n.step + 1) & n.stepMask

        // 如果步进序列达到上限，说明同一毫秒内的 ID 已用尽
        if n.step == 0 {
            // 等待到下一毫秒
            for now <= n.time {
                now = time.Since(n.epoch).Nanoseconds() / 1000000
            }
        }
    } else {
        // 如果是新的时间戳，重置步进序列
        n.step = 0
    }

    // 更新最后生成 ID 的时间戳为当前时间戳
    n.time = now

    // 按照以下顺序生成唯一 ID：
    // 1. 左移时间戳至高位部分 (n.timeShift)，占据 ID 的时间部分
    // 2. 左移节点 ID 至中间部分 (n.nodeShift)，用于区分不同节点
    // 3. 将步进序列作为低位部分，区分同一时间内生成的多个 ID
    r := ID((now)<<n.timeShift | // 时间戳部分
        (n.node << n.nodeShift) | // 节点 ID 部分
        (n.step),                 // 步进序列部分
    )

    // 解锁以允许其他线程访问
    n.mu.Unlock()

    // 返回生成的唯一 ID
    return r
}
```







## 6 问题总结

### 1 这个短链系统可以做过期键吗？怎么做？

### 2 除了base62算法外怎么让其更短？

### 3 雪花算法，处理时钟问题？

https://www.cnblogs.com/thisiswhy/p/17611163.html

### 4.不用时间戳，加1？

https://seata.apache.org/zh-cn/blog/seata-analysis-UUID-generator/

(其实这些也都接近分布式id的一些方法了，有必要了解一下)

### 5.分布式ID有哪些？

美团开元号段模式和SnowFlake。


## 7 最后

仁爱社团成员提供，短链系统。