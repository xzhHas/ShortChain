package interfaces

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/BitofferHub/pkg/middlewares/log"
	pb "github.com/BitofferHub/proto_center/api/shortUrlXsvr/v1"
	"github.com/gin-gonic/gin"
)

func (h *Handler) RedirectToLongUrlV1(ctx *gin.Context) {
	// URL：是完整的地址，包括协议和主机信息
	// URI：关注请求的路径和查询的部分
	fullUrl := ctx.Request.RequestURI
	urlParts := strings.Split(fullUrl, "/v1/")
	if len(urlParts) != 2 || urlParts[1] == "" {
		log.Infof("getLongUrl: 无效的短链 %s", fullUrl)
		ctx.Status(http.StatusNotFound)
		return
	}

	shortURL := urlParts[1]
	reply, err := h.shortUrlXService.GetLongUrlV1(context.Background(), &pb.GetLongUrlRequest{ShortUrl: shortURL})
	if err != nil {
		log.Errorf("controller.RedirectToLongUrlV1 err: %v", err)
		ctx.Status(http.StatusNotFound)
		return
	}

	if reply.LongUrl == "" {
		log.Infof("getLongUrl: 无效的短链 %s", fullUrl)
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, reply.LongUrl)
}

func (h *Handler) CreateShortUrlV1(ctx *gin.Context) {
	var req pb.CreateShortUrlRequest
	if err := ctx.ShouldBind(&req); err != nil {
		log.Errorf("CreateShortUrlV1 err: %v", err.Error())
		ctx.JSON(200, gin.H{
			"message": err.Error(),
		})
		return
	}

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
