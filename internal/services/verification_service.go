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

// VerificationService defines the interface for the verification code service.
type VerificationService interface {
	// GenerateEmailVerificationCode generates and stores an email verification code.
	GenerateEmailVerificationCode(ctx context.Context, email string) (string, error)
	// VerifyEmailVerificationCode verifies an email verification code.
	VerifyEmailVerificationCode(ctx context.Context, email, code string) (bool, error)
	// GeneratePasswordResetToken generates and stores a password reset token.
	GeneratePasswordResetToken(ctx context.Context, email string) (string, error)
	// VerifyPasswordResetToken verifies a password reset token.
	VerifyPasswordResetToken(ctx context.Context, email, token string) (bool, error)
}

// verificationService is the implementation of the VerificationService.
type verificationService struct {
	redis *redis.Client
}

// NewVerificationService creates a new instance of the verification service.
func NewVerificationService(redisClient *redis.Client) VerificationService {
	return &verificationService{
		redis: redisClient,
	}
}

// GenerateEmailVerificationCode generates and stores an email verification code.
func (s *verificationService) GenerateEmailVerificationCode(ctx context.Context, email string) (string, error) {
	// Generate a 6-digit numeric code.
	code, err := s.generateNumericCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	// Store in Redis with an expiration of 10 minutes.
	key := fmt.Sprintf("email_verification:%s", email)
	// ctx := context.Background() // Use passed context

	err = s.redis.Set(ctx, key, code, 10*time.Minute).Err()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store verification code in Redis", "email", email, "error", err) // Use slog.ErrorContext
		return "", fmt.Errorf("failed to store verification code: %w", err)
	}

	slog.InfoContext(ctx, "Email verification code generated", "email", email) // Use slog.InfoContext
	return code, nil
}

// VerifyEmailVerificationCode verifies an email verification code.
func (s *verificationService) VerifyEmailVerificationCode(ctx context.Context, email, code string) (bool, error) {
	key := fmt.Sprintf("email_verification:%s", email)
	// ctx := context.Background() // Use passed context

	storedCode, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.WarnContext(ctx, "Email verification code not found or expired", "email", email) // Use slog.WarnContext
			return false, nil
		}
		slog.ErrorContext(ctx, "Failed to get verification code from Redis", "email", email, "error", err) // Use slog.ErrorContext
		return false, fmt.Errorf("failed to verify code: %w", err)
	}

	// Delete the code after successful verification (one-time use).
	if storedCode == code {
		s.redis.Del(ctx, key)
		slog.InfoContext(ctx, "Email verification code verified successfully", "email", email) // Use slog.InfoContext
		return true, nil
	}

	slog.WarnContext(ctx, "Invalid email verification code", "email", email) // Use slog.WarnContext
	return false, nil
}

// GeneratePasswordResetToken generates and stores a password reset token.
func (s *verificationService) GeneratePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Generate an 8-character alphanumeric token.
	token, err := s.generateAlphanumericToken(8)
	if err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Store in Redis with an expiration of 30 minutes.
	key := fmt.Sprintf("password_reset:%s", email)
	// ctx := context.Background() // Use passed context

	err = s.redis.Set(ctx, key, token, 30*time.Minute).Err()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store password reset token in Redis", "email", email, "error", err) // Use slog.ErrorContext
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	slog.InfoContext(ctx, "Password reset token generated", "email", email) // Use slog.InfoContext
	return token, nil
}

// VerifyPasswordResetToken verifies a password reset token.
func (s *verificationService) VerifyPasswordResetToken(ctx context.Context, email, token string) (bool, error) {
	key := fmt.Sprintf("password_reset:%s", email)
	// ctx := context.Background() // Use passed context

	storedToken, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.WarnContext(ctx, "Password reset token not found or expired", "email", email) // Use slog.WarnContext
			return false, nil
		}
		slog.ErrorContext(ctx, "Failed to get reset token from Redis", "email", email, "error", err) // Use slog.ErrorContext
		return false, fmt.Errorf("failed to verify token: %w", err)
	}

	// Delete the token after successful verification (one-time use).
	if storedToken == token {
		s.redis.Del(ctx, key)
		slog.InfoContext(ctx, "Password reset token verified successfully", "email", email) // Use slog.InfoContext
		return true, nil
	}

	slog.WarnContext(ctx, "Invalid password reset token", "email", email) // Use slog.WarnContext
	return false, nil
}

// generateNumericCode generates a numeric verification code of a given length.
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

// generateAlphanumericToken generates an alphanumeric token of a given length.
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
