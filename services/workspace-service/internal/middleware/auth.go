package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	secret     string
	rdb        *redis.Client
	grpcClient pb.AuthServiceClient
}

func NewAuthMiddleware(secret string, rdb *redis.Client, grpc pb.AuthServiceClient) *AuthMiddleware {
	return &AuthMiddleware{
		secret:     secret,
		rdb:        rdb,
		grpcClient: grpc,
	}
}

func (m *AuthMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secret), nil
		})

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		userID := fmt.Sprintf("%v", claims["userId"])
		deviceId := fmt.Sprintf("%v", claims["deviceId"])
		redisKey := fmt.Sprintf("session:%s:%s", userID, deviceId)

		sessionData, err := m.rdb.Get(context.Background(), redisKey).Result()
		if err == nil && sessionData != "" {
			c.Set("user_id", userID)
			c.Next()
			return
		}

		resp, err := m.grpcClient.VerifySession(context.Background(), &pb.VerifyRequest{
			Token: tokenString,
		})

		if err != nil || !resp.IsValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired", "details": "gRPC check failed"})
			return
		}

		m.rdb.Set(context.Background(), redisKey, "active", 15*time.Minute)
		c.Set("user_id", resp.UserId)
		c.Next()
	}
}
