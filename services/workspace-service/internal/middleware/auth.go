package middleware

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
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
			c.Error(apperror.Unauthorized("Authorization header is required"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secret), nil
		})

		if err != nil || token == nil {
			c.Error(apperror.Unauthorized("Invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.Error(apperror.Unauthorized("Invalid token claims"))
			c.Abort()
			return
		}

		userID := fmt.Sprintf("%v", claims["userId"])
		deviceId := fmt.Sprintf("%v", claims["deviceId"])
		redisKey := fmt.Sprintf("session:%s:%s", userID, deviceId)

		// Check local cache session
		sessionData, err := m.rdb.Get(c.Request.Context(), redisKey).Result()
		if err == nil && sessionData != "" {
			c.Set("user_id", userID)
			c.Next()
			return
		}

		resp, err := m.grpcClient.VerifySession(c.Request.Context(), &pb.VerifyRequest{
			Token: tokenString,
		})

		if err != nil || !resp.IsValid {
			c.Error(apperror.Unauthorized("Session expired or invalid"))
			c.Abort()
			return
		}

		m.rdb.Set(c.Request.Context(), redisKey, "active", 15*time.Minute)
		c.Set("user_id", resp.UserId)
		c.Next()
	}
}
