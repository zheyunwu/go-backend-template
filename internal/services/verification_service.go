package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

// VerificationService 验证码服务接口
type VerificationService interface {
	// 生成并存储邮箱验证码
	GenerateEmailVerificationCode(email string) (string, error)
	// 验证邮箱验证码
	VerifyEmailVerificationCode(email, code string) (bool, error)
	// 生成并存储密码重置令牌
	GeneratePasswordResetToken(email string) (string, error)
	// 验证密码重置令牌
	VerifyPasswordResetToken(email, token string) (bool, error)
}

// verificationService 验证码服务实现
type verificationService struct {
	redis *redis.Client
}

// NewVerificationService 创建验证码服务实例
func NewVerificationService(redisClient *redis.Client) VerificationService {
	return &verificationService{
		redis: redisClient,
	}
}

// GenerateEmailVerificationCode 生成并存储邮箱验证码
func (s *verificationService) GenerateEmailVerificationCode(email string) (string, error) {
	// 生成6位数字验证码
	code, err := s.generateNumericCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	// 存储到 Redis，有效期 10 分钟
	key := fmt.Sprintf("email_verification:%s", email)
	ctx := context.Background()

	err = s.redis.Set(ctx, key, code, 10*time.Minute).Err()
	if err != nil {
		slog.Error("Failed to store verification code in Redis", "email", email, "error", err)
		return "", fmt.Errorf("failed to store verification code: %w", err)
	}

	slog.Info("Email verification code generated", "email", email)
	return code, nil
}

// VerifyEmailVerificationCode 验证邮箱验证码
func (s *verificationService) VerifyEmailVerificationCode(email, code string) (bool, error) {
	key := fmt.Sprintf("email_verification:%s", email)
	ctx := context.Background()

	storedCode, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Warn("Email verification code not found or expired", "email", email)
			return false, nil
		}
		slog.Error("Failed to get verification code from Redis", "email", email, "error", err)
		return false, fmt.Errorf("failed to verify code: %w", err)
	}

	// 验证成功后删除验证码（一次性使用）
	if storedCode == code {
		s.redis.Del(ctx, key)
		slog.Info("Email verification code verified successfully", "email", email)
		return true, nil
	}

	slog.Warn("Invalid email verification code", "email", email)
	return false, nil
}

// GeneratePasswordResetToken 生成并存储密码重置令牌
func (s *verificationService) GeneratePasswordResetToken(email string) (string, error) {
	// 生成8位字母数字混合令牌
	token, err := s.generateAlphanumericToken(8)
	if err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}

	// 存储到 Redis，有效期 30 分钟
	key := fmt.Sprintf("password_reset:%s", email)
	ctx := context.Background()

	err = s.redis.Set(ctx, key, token, 30*time.Minute).Err()
	if err != nil {
		slog.Error("Failed to store password reset token in Redis", "email", email, "error", err)
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	slog.Info("Password reset token generated", "email", email)
	return token, nil
}

// VerifyPasswordResetToken 验证密码重置令牌
func (s *verificationService) VerifyPasswordResetToken(email, token string) (bool, error) {
	key := fmt.Sprintf("password_reset:%s", email)
	ctx := context.Background()

	storedToken, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Warn("Password reset token not found or expired", "email", email)
			return false, nil
		}
		slog.Error("Failed to get reset token from Redis", "email", email, "error", err)
		return false, fmt.Errorf("failed to verify token: %w", err)
	}

	// 验证成功后删除令牌（一次性使用）
	if storedToken == token {
		s.redis.Del(ctx, key)
		slog.Info("Password reset token verified successfully", "email", email)
		return true, nil
	}

	slog.Warn("Invalid password reset token", "email", email)
	return false, nil
}

// generateNumericCode 生成数字验证码
func (s *verificationService) generateNumericCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += num.String()
	}
	return code, nil
}

// generateAlphanumericToken 生成字母数字混合令牌
func (s *verificationService) generateAlphanumericToken(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, length)

	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[num.Int64()]
	}

	return string(token), nil
}
