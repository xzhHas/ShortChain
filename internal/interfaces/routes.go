package interfaces

import (
	"net/http"

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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func NewRouter(h *Handler) *gin.Engine {
	r := engine.NewEngine(engine.WithLogger(false))
	r.Use(corsMiddleware())

	project := r.Group("/api")
	project.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	/*v3版本*
	 *1.雪花算法生成自增ID*
	 *2.base62算法*
	 *3.Redis缓存来加速查询长链和短链*
	 */
	//v3 := project.Group("/v3")
	{
		project.GET("/:short_url", h.RedirectToLongUrlV3)
		project.POST("/shorten", h.CreateShortUrlV3)
	}

	/*v3版本*
	 *1.雪花算法生成自增ID*
	 *2.base62算法*
	 *3.Redis缓存来加速查询长链和短链*
	 */
	//v3 := project.Group("/v3")
	{
		r.GET("/:short_url", h.RedirectToLongUrlV3)
		r.POST("/shorten", h.CreateShortUrlV3)
	}

	/*
		v1版本
		单纯依赖数据库获取自增ID，通过base62转换成短链
	*/
	// v1 := project.Group("/v1")
	// {
	// 	v1.GET("/:short_url", h.RedirectToLongUrlV1)
	// 	v1.POST("/shorten", h.CreateShortUrlV1)
	// }
	/*
		v2版本
		1.依赖于Redis，自增ID的性能
		2.base62算法将获得的自增ID转换成短链
		3.利用Redis布隆过滤器，拦截一部分访问DB的流量，可以做快速判断
	*/
	// v2 := project.Group("/v2")
	// {
	// 	v2.GET("/:short_url", h.RedirectToLongUrlV2)
	// 	v2.POST("/shorten", h.CreateShortUrlV2)
	// }
	return r
}
