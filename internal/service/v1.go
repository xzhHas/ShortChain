package service

import (
	"context"

	pb "github.com/BitofferHub/proto_center/api/shortUrlXsvr/v1"
)

func (s *ShortUrlXService) GetLongUrlV1(ctx context.Context, req *pb.GetLongUrlRequest) (*pb.GetLongUrlReply, error) {
	longUrl, err := s.uc.GetLongUrlV1(ctx, req.ShortUrl)
	if err != nil {
		return nil, err
	}
	return &pb.GetLongUrlReply{LongUrl: longUrl}, nil
}

func (s *ShortUrlXService) GenerateShortUrlV1(ctx context.Context, req *pb.CreateShortUrlRequest) (*pb.CreateShortUrlReply, error) {
	// 应用层调用，主要负责协调输入请求和底层逻辑的处理
	shortUrl, err := s.uc.GenerateShortUrlV1(ctx, req.LongUrl)
	if err != nil {
		return nil, err
	}

	return &pb.CreateShortUrlReply{ShortUrl: shortUrl}, nil
}
