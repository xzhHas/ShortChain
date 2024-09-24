package interfaces

import (
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	pb "github.com/BitofferHub/proto_center/api/shortUrlXsvr/v1"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

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
