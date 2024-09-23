package service

import (
	pb "github.com/BitofferHub/proto_center/api/shortUrlXsvr/v1"
	"github.com/BitofferHub/shortUrlX/internal/biz"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewShortUrlXService)

type ShortUrlXService struct {
	pb.UnimplementedShortUrlXServer
	uc *biz.UrlMapUseCase // 依赖注入了biz的接口
}

// 已传参的形式，依赖注入
func NewShortUrlXService(uc *biz.UrlMapUseCase) *ShortUrlXService {
	return &ShortUrlXService{uc: uc}
}
