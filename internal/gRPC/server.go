package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	uc "github.com/SerzhLimon/ReductionURL/internal/service"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const secretKey = "super_secret_key_for_shortener"

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type ShortenerServer struct {
	UnimplementedShortenerServiceServer
	uc  uc.UseCase
	cfg *config.Config
}

func NewShortenerServer(useCase uc.UseCase, cfg *config.Config) *ShortenerServer {
	return &ShortenerServer{
		uc:  useCase,
		cfg: cfg,
	}
}

func (s *ShortenerServer) getUserIDFromContext(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if tokenStr == authHeaders[0] {
		return 0, status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	if claims.UserID < 1 {
		return 0, status.Errorf(codes.Unauthenticated, "invalid user id")
	}
	return claims.UserID, nil
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, req *URLShortenRequest) (*URLShortenResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.uc.SetUser(userID)

	hash, err := s.uc.SetURL(req.Url)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			fullURL := s.cfg.Opts.BaseURL + "/" + hash
			return &URLShortenResponse{Result: fullURL}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to shorten url: %v", err)
	}

	fullURL := s.cfg.Opts.BaseURL + "/" + hash
	return &URLShortenResponse{Result: fullURL}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *URLExpandRequest) (*URLExpandResponse, error) {
	// 1. Получить оригинальный URL по id
	originalURL, err := s.uc.GetURL(req.Id)
	if err != nil {
		if errors.Is(err, model.ErrDeletedURL) {
			return nil, status.Errorf(codes.NotFound, "url not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get url: %v", err)
	}

	// 2. Вернуть результат
	return &URLExpandResponse{Result: originalURL}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*UserURLsResponse, error) {
	// 1. Получить userID из контекста
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	// 2. Установить пользователя
	s.uc.SetUser(userID)
	
	// 3. Получить список URL пользователя
	urls, err := s.uc.GetArrayURL()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user urls: %v", err)
	}
	
	// 4. Преобразовать в gRPC ответ
	pbURLs := make([]*URLData, len(urls))
	for i, u := range urls {
		pbURLs[i] = &URLData{
			ShortUrl:    s.cfg.Opts.BaseURL + "/" + u.Short,
			OriginalUrl: u.Original,
		}
	}
	
	return &UserURLsResponse{Urls: pbURLs}, nil
}
