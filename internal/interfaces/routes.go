// routes.go

package interfaces

import (
	engine "github.com/BitofferHub/pkg/middlewares/gin"
	"github.com/BitofferHub/shortUrlX/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	shortUrlXService *service.ShortUrlXService
}

func NewHandler(s *service.ShortUrlXService) *Handler {
	return &Handler{
		shortUrlXService: s,
	}
}

func NewRouter(h *Handler) *gin.Engine {
	r := engine.NewEngine(engine.WithLogger(false))

	project := r.Group("/shorturlx")

	project.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	/*
		v1版本
		单纯依赖数据库获取自增ID，通过base62转换成短链
	*/
	v1 := project.Group("/v1")
	{
		// Get 传入短链，返回长链
		v1.GET("/:short_url", h.RedirectToLongUrlV1)
		// 单纯依赖数据库生成短链
		// Post 创建长链对应短链（ps：我们做的是一一对应的，一个长链只对应一个短链）
		v1.POST("/shorten", h.CreateShortUrlV1)
	}
	/*
		v2版本
		1.依赖于Redis，自增ID的性能
		2.base62算法将获得的自增ID转换成短链
		3.利用Redis布隆过滤器，拦截一部分访问DB的流量，可以做快速判断
	*/
	v2 := project.Group("/v2")
	{
		// Get 传入短链，返回长链
		v2.GET("/:short_url", h.RedirectToLongUrlV2)

		// 使用Redis来生成短链
		// Post 创建长链对应短链（ps：我们做的是一一对应的，一个长链只对应一个短链）
		v2.POST("/shorten", h.CreateShortUrlV2)
	}
	/*
		v3版本
		1.雪花算法生成自增ID
		2.base62算法
		3.Redis缓存来加速查询长链和短链
	*/
	v3 := project.Group("/v3")
	{
		// Get 传入短链，返回长链
		v3.GET("/:short_url", h.RedirectToLongUrlV3)

		// 使用雪花算法来生成短链
		// Post 创建长链对应短链（ps：我们做的是一一对应的，一个长链只对应一个短链）
		v3.POST("/shorten", h.CreateShortUrlV3)
	}
	return r
}
