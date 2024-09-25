# 短链系统

## 认识短链

### 1.什么是短链？为什么需要短链？

短链的优点：

1. 更加简洁
2. 便于使用
   1. 短链接生成的二维码更好识别
3. 节省成本
   1. 短信按长度收费，短链接节省成本

### 2.短链的原理是什么？

短链实际上就是：访问一个极短的地址，然后重定向到原始地址访问。

![image-20240924190123215](C:\Users\xz317\AppData\Roaming\Typora\typora-user-images\image-20240924190123215.png)

### 3.如何生成短链？

- 哈希算法
- 唯一ID
  - MySQL自增ID
  - Redis自增ID
  - Snowflake 雪花算法

## 短链设计

### 1.重定向状态码的选择

永久重定向：301

临时重定向：302



### 2.接口的设计

| HTTP请求方法 | 路径          | 请求体数据 | 响应                    |
| ------------ | ------------- | ---------- | ----------------------- |
| GET          | /:{short_url} | null       | 状态码：302，临时重定向 |
| POST         | /shorten      | long_url   | short_url               |



### 3.持久化存储

- MySQL
- Redis
  - 长链到短链的映射
  - 短链到长链的映射



## 短链实践

### 1.生成短链？

#### v1版本

- 生成短链依赖MySQL获取一个自增ID
- 通过base62算法将ID转换成短链

/internal/interfacess/v1.go

```go
func (h *Handler) CreateShortUrlV1(ctx *gin.Context) {
	// kratos 框架集成 gin框架用法，使用gin.Context
	// 这里使用微服务，使用protobuf，有用于grpc通信的Request和回复Reply
	var req pb.CreateShortUrlRequest
	// 将客户端发送的json数据，ShouldBind给req
	if err := ctx.ShouldBind(&req); err != nil {
		log.Errorf("CreateShortUrlV1 err: %v", err.Error())
		ctx.JSON(200, gin.H{
			"message": err.Error(),
		})
		return
	}
	// 通过grpc调用后端服务
	reply, err := h.shortUrlXService.GenerateShortUrlV1(context.Background(), &req)
	if err != nil {
		log.Errorf("CreateShortUrlV1 err: %v", err)
		ctx.JSON(200, gin.H{
			"message": err.Error(),
		})
		return
	}
	
	ctx.JSON(200, pb.CreateShortUrlReply{ShortUrl: fmt.Sprintf("http://%s/v1/%s", ctx.Request.Host, reply.ShortUrl)})
}

func (h *Handler) RedirectToLongUrlV1(ctx *gin.Context) {
	// URL：是完整的地址，包括协议和主机信息
	// URI：关注请求的路径和查询的部分
	fullUrl := ctx.Request.RequestURI // 获取完整的请求URL，也就是获取路由的路径
	//fmt.Printf("fullurl：%s\n", fullUrl)
	urlParts := strings.Split(fullUrl, "/v1/")
	///shorturlx/v1/y
	//fmt.Println(urlParts) /shorturlx y
	if len(urlParts) != 2 || urlParts[1] == "" {
		log.Infof("getLongUrl: 无效的短链 %s", fullUrl)
		ctx.Status(http.StatusNotFound)
		return
	}

	shortURL := urlParts[1]
	//fmt.Println(shortURL)  y
	reply, err := h.shortUrlXService.GetLongUrlV1(context.Background(), &pb.GetLongUrlRequest{ShortUrl: shortURL})
	if err != nil {
		log.Errorf("controller.RedirectToLongUrlV1 err: %v", err)
		ctx.Status(http.StatusNotFound)
		return
	}

	// 不存在，返回404
	if reply.LongUrl == "" {
		log.Infof("getLongUrl: 无效的短链 %s", fullUrl)
		ctx.Status(http.StatusNotFound)
		return
	}

	// 存在，重定向到对应短链，301
	ctx.Redirect(http.StatusMovedPermanently, reply.LongUrl)
}

// GetLongUrlV1 调用业务逻辑层，获取长链
// @return LongUrl
func (s *ShortUrlXService) GetLongUrlV1(ctx context.Context, req *pb.GetLongUrlRequest) (*pb.GetLongUrlReply, error) {
	longUrl, err := s.uc.GetLongUrlV1(ctx, req.ShortUrl)
	if err != nil {
		return nil, err
	}
	return &pb.GetLongUrlReply{LongUrl: longUrl}, nil
}
```

/internale/service/v1.go

```go
func (s *ShortUrlXService) GenerateShortUrlV1(ctx context.Context, req *pb.CreateShortUrlRequest) (*pb.CreateShortUrlReply, error) {
	// 应用层调用，主要负责协调输入请求和底层逻辑的处理
	shortUrl, err := s.uc.GenerateShortUrlV1(ctx, req.LongUrl)
	if err != nil {
		return nil, err
	}

	return &pb.CreateShortUrlReply{ShortUrl: shortUrl}, nil
}

// GetLongUrlV1 获取长链
func (uc *UrlMapUseCase) GetLongUrlV1(ctx context.Context, shortUrl string) (string, error) {
	return uc.repo.GetLongUrlFormDb(ctx, shortUrl)
}

```

/internal/biz/v1.go (主要逻辑)

![image-20240924195806240](C:\Users\xz317\AppData\Roaming\Typora\typora-user-images\image-20240924195806240.png)

```go
// GenerateShortUrlV1 领域，业务逻辑层
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
```

  

**总结**

1、存储短链

- 输入long_url，如果存在对应的short_url，则返回；
- 将long_url存储到mysql中，并返回这个数据的ID；
- 通过本地base62算法，将ID转换成short_url；
- 更新存储long_url信息的short_url字段。

2、查询短链

- 通过short_url查询对应的long_url，如果存在就重定向到long_url，否则返回nil。

#### v2版本

- 依赖Redis生成自增ID（incr命令），提升性能；
- base62将获得的自增ID转换成短链；
- 利用Redis布隆过滤器，拦截一部分DB访问量，可以快速判断；

**生成短链**

![image-20240924205210709](C:\Users\xz317\AppData\Roaming\Typora\typora-user-images\image-20240924205210709.png)

**查询短链**

![image-20240925122438104](C:\Users\xz317\AppData\Roaming\Typora\typora-user-images\image-20240925122438104.png)



#### v3版本（重点）

- 通过本地雪花算法获取自增ID
- base62算法将获得的自增ID转换成短链
- 用Redis缓存来加速查询长链和短链

**生成短链**

![image-20240925160426359](C:\Users\xz317\AppData\Roaming\Typora\typora-user-images\image-20240925160426359.png)



**查询短链**

![image-20240925161101429](C:\Users\xz317\AppData\Roaming\Typora\typora-user-images\image-20240925161101429.png)











## 短链系统总结

### 1、base62算法

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
```

### 2、布隆过滤器

布隆过滤器是用来检索一个元素是否存在于一个集合中。

布隆过滤器是由很长的二进制向量与一系列随机映射函数构成。

使用场景

- Redis通过布隆过滤器防止缓存穿透
- RocketMQ通过布隆过滤器防止消息重复消费

