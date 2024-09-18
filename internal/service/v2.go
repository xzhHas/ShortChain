package service

import (
	"context"
	pb "github.com/BitofferHub/proto_center/api/shortUrlXsvr/v1"
)

func (s *ShortUrlXService) GenerateShortUrlV2(ctx context.Context, req *pb.CreateShortUrlRequest) (*pb.CreateShortUrlReply, error) {
	shortUrl, err := s.uc.GenerateShortUrlV2(ctx, req.LongUrl)
	if err != nil {
		return nil, err
	}

	return &pb.CreateShortUrlReply{ShortUrl: shortUrl}, nil
}
func (s *ShortUrlXService) GetLongUrlV2(ctx context.Context, req *pb.GetLongUrlRequest) (*pb.GetLongUrlReply, error) {
	longUrl, err := s.uc.GetLongUrlV2(ctx, req.ShortUrl)
	if err != nil {
		return nil, err
	}
	return &pb.GetLongUrlReply{LongUrl: longUrl}, nil
}
