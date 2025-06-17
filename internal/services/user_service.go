package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/jwt"
	"github.com/go-backend-template/pkg/logger"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	/* Admin-facing business logic */
	ListUsers(ctx context.Context, params *query_params.QueryParams, includeSoftDeleted ...bool) ([]models.User, *response.Pagination, error)
	GetUser(ctx context.Context, id uint, includeSoftDeleted ...bool) (*models.User, error)
	CreateUser(ctx context.Context, req *dto.RegisterWithPasswordRequest) (uint, error)
	UpdateUser(ctx context.Context, id uint, req *dto.UpdateProfileRequest) error
	DeleteUser(ctx context.Context, id uint) error
	RestoreUser(ctx context.Context, id uint) error          // Restore soft-deleted user
	BanUser(ctx context.Context, id uint, banned bool) error // Ban or unban user

	/* Auth logic */
	UpdatePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error

	/* Traditional registration/login related */
	RegisterWithPassword(ctx context.Context, req *dto.RegisterWithPasswordRequest) (uint, error)
	LoginWithPassword(ctx context.Context, emailOrPhone, password string) (string, string, error) // Returns (accessToken, refreshToken, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error)          // Returns (newAccessToken, newRefreshToken, error)

	/* Email verification related */
	SendEmailVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, email, code string) error

	/* Password reset related */
	SendPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, resetToken, newPassword string) error

	/* WeChat Mini Program related */
	RegisterFromWechatMiniProgram(ctx context.Context, req *dto.RegisterFromWechatMiniProgramRequest, unionID *string, openID *string) (uint, error)
	LoginFromWechatMiniProgram(ctx context.Context, unionID *string, openID *string) (string, string, error) // Returns (accessToken, refreshToken, error)

	/* WeChat OAuth2.0 (App/Web) related */
	ExchangeWechatOAuth(ctx context.Context, req *dto.WechatOAuthRequest) (string, string, bool, error) // Returns (accessToken, refreshToken, isNewUser, error)
	BindWechatAccount(ctx context.Context, userID uint, req *dto.BindWechatAccountRequest, authenticatedUser *models.User) error
	UnbindWechatAccount(ctx context.Context, userID uint, authenticatedUser *models.User) error
	/* Google OAuth2.0 (App/Web) related */
	ExchangeGoogleOAuth(ctx context.Context, req *dto.GoogleOAuthRequest) (string, string, bool, error) // Returns (accessToken, refreshToken, isNewUser, error)
	BindGoogleAccount(ctx context.Context, userID uint, req *dto.BindGoogleAccountRequest, authenticatedUser *models.User) error
	UnbindGoogleAccount(ctx context.Context, userID uint, authenticatedUser *models.User) error
}

// userService is the implementation of UserService.
type userService struct {
	config              *config.Config
	userRepo            repositories.UserRepository
	googleOAuthService  GoogleOAuthService
	emailService        EmailService
	verificationService VerificationService
}

// NewUserService creates a new instance of UserService.
func NewUserService(config *config.Config, userRepo repositories.UserRepository, googleOAuthService GoogleOAuthService, emailService EmailService, verificationService VerificationService) UserService {
	return &userService{
		config:              config,
		userRepo:            userRepo,
		googleOAuthService:  googleOAuthService,
		emailService:        emailService,
		verificationService: verificationService,
	}
}

/*
Admin-facing business logic
*/

// ListUsers retrieves a list of users.
func (s *userService) ListUsers(ctx context.Context, params *query_params.QueryParams, includeSoftDeleted ...bool) ([]models.User, *response.Pagination, error) {
	// Call the repository layer to get the list of users.
	userList, total, err := s.userRepo.ListUsers(ctx, params, includeSoftDeleted...) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to list users", "error", err) // Use slog.ErrorContext
		return nil, nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Return an empty array if there is no data.
	if len(userList) == 0 {
		userList = []models.User{}
	}

	// Construct pagination information.
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return userList, pagination, nil
}

// GetUser retrieves details for a single user.
func (s *userService) GetUser(ctx context.Context, id uint, includeSoftDeleted ...bool) (*models.User, error) {
	// Call the repository layer to get the user.
	user, err := s.userRepo.GetUser(ctx, id, includeSoftDeleted...) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to get user from repository", "userId", id, "error", err) // Use slog.ErrorContext
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser creates a user (traditional Email/Phone + password).
func (s *userService) CreateUser(ctx context.Context, req *dto.RegisterWithPasswordRequest) (uint, error) {
	// Validate that at least email or phone is provided.
	if req.Email == "" {
		return 0, errors.ErrEmailNotProvided
	}

	// Validate for duplicate email or phone number.
	if req.Email != "" {
		if _, err := s.userRepo.GetUserByField(ctx, "email", req.Email, true); err == nil { // Pass context
			return 0, errors.ErrEmailAlreadyExists
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if _, err := s.userRepo.GetUserByField(ctx, "phone", *req.Phone, true); err == nil { // Pass context
			return 0, errors.ErrPhoneAlreadyExists
		}
	}

	// Hash the password.
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		logger.Error(ctx, "Failed to hash password", "error", err) // Use slog.ErrorContext
		return 0, fmt.Errorf("error processing password: %w", err)
	}
	// Convert DTO to User model.
	user := req.ToModel(hashedPassword)

	// Call the repository layer to create the user.
	if err := s.userRepo.CreateUser(ctx, user); err != nil { // Pass context
		logger.Error(ctx, "Failed to create user", // Use slog.ErrorContext
			"name", user.Name,
			"email", req.Email,
			"phone", req.Phone,
			"error", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// If the user provided an email, automatically send a verification email.
	if req.Email != "" {
		// Generate verification code.
		code, err := s.verificationService.GenerateEmailVerificationCode(ctx, req.Email) // Pass context
		if err != nil {
			logger.Warn(ctx, "Failed to generate email verification code after user creation", "userId", user.ID, "email", req.Email, "error", err) // Use slog.WarnContext
			// Do not return an error here, as the user has already been created successfully; only email verification failed.
		} else {
			// Send verification email, using the user's language preference.
			err = s.emailService.SendEmailVerification(ctx, req.Email, user.Name, code, user.Locale) // Pass context
			if err != nil {
				logger.Warn(ctx, "Failed to send verification email after user creation", "userId", user.ID, "email", req.Email, "error", err) // Use slog.WarnContext
				// Do not return an error here, as the user has already been created successfully.
			} else {
				logger.Info(ctx, "Verification email sent successfully after user creation", "userId", user.ID, "email", req.Email, "language", user.Locale) // Use slog.InfoContext
			}
		}
	}

	return user.ID, nil
}

// hashPassword hashes a password using bcrypt.
func hashPassword(password string) (string, error) {
	// Generate a hash with the recommended cost (10-12).
	// bcrypt automatically adds a random salt and includes it in the hash result.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// UpdateUser updates a user.
func (s *userService) UpdateUser(ctx context.Context, id uint, req *dto.UpdateProfileRequest) error {
	// Check if the user exists.
	user, err := s.GetUser(ctx, id) // Pass context
	if err != nil {
		return err
	}

	// Convert DTO to a map of fields to update.
	updates := req.ToUpdatesMap()

	// If Email is updated, check if it already exists and set it to unverified.
	if req.Email != nil && *req.Email != "" && (user.Email == nil || *user.Email != *req.Email) {
		existingUser, err := s.userRepo.GetUserByField(ctx, "email", *req.Email) // Pass context
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error(ctx, "Failed to check existing email", "email", *req.Email, "error", err) // Use slog.ErrorContext
			return fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingUser != nil && existingUser.ID != 0 && existingUser.ID != id {
			logger.Warn(ctx, "Email already exists", "email", *req.Email) // Use slog.WarnContext
			return errors.ErrEmailAlreadyExists
		}
		// If Email is updated, set it to unverified.
		updates["is_email_verified"] = false
	}

	// Call the repository layer to perform the update.
	if err := s.userRepo.UpdateUser(ctx, id, updates); err != nil { // Pass context
		logger.Error(ctx, "Failed to update user", "userId", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user.
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	// Check if the user exists.
	_, err := s.GetUser(ctx, id) // Pass context
	if err != nil {
		return err
	}

	// Call the repository layer to delete the user.
	if err := s.userRepo.DeleteUser(ctx, id); err != nil { // Pass context
		logger.Error(ctx, "Failed to delete user", "userId", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// RestoreUser restores a soft-deleted user.
func (s *userService) RestoreUser(ctx context.Context, id uint) error {
	// Check if the user exists (including soft-deleted).
	user, err := s.GetUser(ctx, id, true) // Include soft-deleted records, Pass context
	if err != nil {
		return err
	}

	// If the user is not soft-deleted, return directly.
	if !user.DeletedAt.Valid {
		return nil
	}

	// Construct a map of fields to update.
	updates := map[string]interface{}{"deleted_at": nil}

	// Call the repository layer to restore the user.
	if err := s.userRepo.UpdateUser(ctx, id, updates, true); err != nil { // Pass context
		logger.Error(ctx, "Failed to restore user", "userId", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to restore user: %w", err)
	}

	return nil
}

// BanUser bans or unbans a user.
func (s *userService) BanUser(ctx context.Context, id uint, isBanned bool) error {
	// Check if the user exists.
	_, err := s.GetUser(ctx, id) // Pass context
	if err != nil {
		return err
	}

	// Construct a map of fields to update.
	updates := map[string]interface{}{
		"is_banned": isBanned,
	}

	// Call the repository layer to update the user's ban status.
	if err := s.userRepo.UpdateUser(ctx, id, updates); err != nil { // Pass context
		logger.Error(ctx, "Failed to update ban status", "userId", id, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to update ban status: %w", err)
	}
	return nil
}

/*
Auth logic
*/

// UpdatePassword updates a user's password.
func (s *userService) UpdatePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	// Check if the user exists.
	user, err := s.userRepo.GetUser(ctx, userID) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to find user", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("database error: %w", err)
	}

	// Validate the current password.
	if user.Password == nil || *user.Password == "" {
		logger.Warn(ctx, "User has no password set", "userId", user.ID) // Use slog.WarnContext
		return errors.ErrInvalidPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword))
	if err != nil {
		logger.Warn(ctx, "Current password verification failed", "userId", user.ID) // Use slog.WarnContext
		return errors.ErrInvalidPassword
	}

	// Hash the new password.
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		logger.Error(ctx, "Failed to hash new password", "userId", user.ID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the user's password.
	updates := map[string]interface{}{
		"password": hashedPassword,
	}

	// Call the repository layer to update the user's password.
	if err := s.userRepo.UpdateUser(ctx, user.ID, updates); err != nil { // Pass context
		logger.Error(ctx, "Failed to update user password", "userId", user.ID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info(ctx, "Password updated successfully", "userId", user.ID) // Use slog.InfoContext
	return nil
}

/*
Traditional registration/login related
*/

// RegisterWithPassword registers a user with a password.
func (s *userService) RegisterWithPassword(ctx context.Context, req *dto.RegisterWithPasswordRequest) (uint, error) {
	// Delegate to CreateUser.
	return s.CreateUser(ctx, req) // Pass context
}

// LoginWithPassword validates user password and generates JWT tokens.
func (s *userService) LoginWithPassword(ctx context.Context, emailOrPhone, password string) (string, string, error) {
	var user *models.User
	var err error

	// Try finding the user by email first.
	user, err = s.userRepo.GetUserByField(ctx, "email", emailOrPhone) // Pass context
	if err != nil && err == gorm.ErrRecordNotFound {
		// If not found by email, try finding by phone number.
		user, err = s.userRepo.GetUserByField(ctx, "phone", emailOrPhone) // Pass context
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return "", "", errors.ErrUserNotFound
			}
			logger.Error(ctx, "Failed to find user by phone", "emailOrPhone", emailOrPhone, "error", err) // Use slog.ErrorContext
			return "", "", fmt.Errorf("database error: %w", err)
		}
	} else if err != nil {
		logger.Error(ctx, "Failed to find user by email", "emailOrPhone", emailOrPhone, "error", err) // Use slog.ErrorContext
		return "", "", fmt.Errorf("database error: %w", err)
	}

	// Check if the user is banned.
	if user.IsBanned {
		return "", "", errors.ErrUserBanned
	}

	// Check if the user has a password set.
	if user.Password == nil || *user.Password == "" {
		logger.Warn(ctx, "User has no password set", "userId", user.ID) // Use slog.WarnContext
		return "", "", errors.ErrInvalidPassword
	}

	// Validate the password.
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		logger.Warn(ctx, "Password verification failed", "userId", user.ID) // Use slog.WarnContext
		return "", "", errors.ErrInvalidPassword
	}

	// Generate JWT access token.
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate access token", "error", err, "userId", user.ID) // Use slog.ErrorContext
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate JWT refresh token.
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate refresh token", "error", err, "userId", user.ID) // Use slog.ErrorContext
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login time.
	now := time.Now()
	s.userRepo.UpdateUser(ctx, user.ID, map[string]interface{}{"last_login": now}) // Pass context

	return accessToken, refreshToken, nil
}

// RefreshAccessToken uses a refresh token to get a new access token.
func (s *userService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 1. Validate the refresh token.
	refreshTokenDetails, err := jwt.ValidateToken(refreshToken, s.config.JWT.Secret)
	if err != nil || refreshTokenDetails.TokenType != jwt.RefreshToken {
		logger.Debug(ctx, "Refresh token validation failed", "error", err) // Use slog.DebugContext
		return "", "", errors.ErrInvalidToken
	}

	// 2. Validate if the user exists and is not banned.
	user, err := s.userRepo.GetUser(ctx, refreshTokenDetails.UserID) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn(ctx, "User not found for refresh token", "userId", refreshTokenDetails.UserID) // Use slog.WarnContext
			return "", "", errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to get user during token refresh", "userId", refreshTokenDetails.UserID, "error", err) // Use slog.ErrorContext
		return "", "", fmt.Errorf("database error: %w", err)
	}

	// 3. Check if the user is banned.
	if user.IsBanned {
		logger.Warn(ctx, "Refresh token rejected for banned user", "userId", user.ID) // Use slog.WarnContext
		return "", "", errors.ErrUserBanned
	}

	// 4. Generate a new access token.
	newAccessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate new access token", "userId", user.ID, "error", err) // Use slog.ErrorContext
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// 5. Generate a new refresh token (optional, for refresh token rotation).
	newRefreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate new refresh token", "userId", user.ID, "error", err) // Use slog.ErrorContext
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 6. Update last login time.
	now := time.Now()
	s.userRepo.UpdateUser(ctx, user.ID, map[string]interface{}{"last_login": now}) // Pass context

	logger.Info(ctx, "Access token refreshed successfully", "userId", user.ID) // Use slog.InfoContext
	return newAccessToken, newRefreshToken, nil
}

/*
Email verification related
*/

// SendEmailVerification sends an email verification code.
func (s *userService) SendEmailVerification(ctx context.Context, email string) error {
	// Check if the user exists.
	user, err := s.userRepo.GetUserByField(ctx, "email", email) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to find user by email", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("database error: %w", err)
	}

	// Check if the email is already verified.
	if user.IsEmailVerified {
		return errors.ErrEmailAlreadyVerified
	}

	// Generate verification code.
	code, err := s.verificationService.GenerateEmailVerificationCode(ctx, email) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to generate email verification code", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to generate verification code: %w", err)
	}

	// Send verification email.
	err = s.emailService.SendEmailVerification(ctx, email, user.Name, code, user.Locale) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to send verification email", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	logger.Info(ctx, "Email verification sent successfully", "email", email) // Use slog.InfoContext
	return nil
}

// VerifyEmail verifies an email.
func (s *userService) VerifyEmail(ctx context.Context, email, code string) error {
	// Check if the user exists.
	user, err := s.userRepo.GetUserByField(ctx, "email", email) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to find user by email", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("database error: %w", err)
	}

	// Check if the email is already verified.
	if user.IsEmailVerified {
		return errors.ErrEmailAlreadyVerified
	}

	// Validate the verification code.
	isValid, err := s.verificationService.VerifyEmailVerificationCode(ctx, email, code) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to verify email verification code", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to verify code: %w", err)
	}

	if !isValid {
		return errors.ErrInvalidVerificationCode
	}

	// Update the user's email verification status.
	updates := map[string]interface{}{
		"is_email_verified": true,
	}

	if err := s.userRepo.UpdateUser(ctx, user.ID, updates); err != nil { // Pass context
		logger.Error(ctx, "Failed to update email verification status", "userId", user.ID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	logger.Info(ctx, "Email verified successfully", "email", email, "userId", user.ID) // Use slog.InfoContext
	return nil
}

/*
Password reset related
*/

// SendPasswordReset sends a password reset email.
func (s *userService) SendPasswordReset(ctx context.Context, email string) error {
	// Check if the user exists.
	user, err := s.userRepo.GetUserByField(ctx, "email", email) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to find user by email", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("database error: %w", err)
	}

	// Generate reset token.
	token, err := s.verificationService.GeneratePasswordResetToken(ctx, email) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to generate password reset token", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Send reset email.
	err = s.emailService.SendPasswordReset(ctx, email, user.Name, token, user.Locale) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to send password reset email", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	logger.Info(ctx, "Password reset email sent successfully", "email", email) // Use slog.InfoContext
	return nil
}

// ResetPassword resets a user's password.
func (s *userService) ResetPassword(ctx context.Context, email, resetToken, newPassword string) error {
	// Check if the user exists.
	user, err := s.userRepo.GetUserByField(ctx, "email", email) // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		logger.Error(ctx, "Failed to find user by email", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("database error: %w", err)
	}

	// Validate the reset token.
	isValid, err := s.verificationService.VerifyPasswordResetToken(ctx, email, resetToken) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to verify password reset token", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to verify reset token: %w", err)
	}

	if !isValid {
		return errors.ErrInvalidVerificationCode
	}

	// Hash the new password.
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		logger.Error(ctx, "Failed to hash new password", "email", email, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the user's password.
	updates := map[string]interface{}{
		"password": hashedPassword,
	}

	if err := s.userRepo.UpdateUser(ctx, user.ID, updates); err != nil { // Pass context
		logger.Error(ctx, "Failed to update user password", "userId", user.ID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info(ctx, "Password reset successfully", "email", email, "userId", user.ID) // Use slog.InfoContext
	return nil
}

/*
WeChat Mini Program related
*/

// RegisterUserFromWechatMiniProgram registers a user from WeChat Mini Program.
func (s *userService) RegisterFromWechatMiniProgram(ctx context.Context, req *dto.RegisterFromWechatMiniProgramRequest, unionID *string, openID *string) (uint, error) {
	// Check sensitive content in production environment.
	// if openID != nil && *openID != "" {
	// 	// Check if user nickname contains sensitive content.
	// 	if err := utils.CheckSensitiveContent(req.Name, *openID, utils.SecuritySceneProfile); err != nil {
	// 		return 0, err
	// 	}
	// }

	// Validate if openID is provided.
	if openID == nil || *openID == "" {
		return 0, errors.ErrOpenIDNotProvided
	}

	// Validate: if the user has already registered via WeChat Mini Program.
	if _, err := s.userRepo.GetUserByProvider(ctx, "wechat_mini_program", *openID); err == nil { // Pass context
		return 0, errors.ErrUserAlreadyExists
	}

	// If UnionID exists, check if the user has registered via APP; if so, use its associated UserID.
	if unionID != nil && *unionID != "" {
		existingUser, err := s.userRepo.GetUserByUnionID(ctx, *unionID) // Pass context
		if err == nil && existingUser != nil && existingUser.ID > 0 {
			// User has already registered via APP, use this user's ID.
			userProvider := models.UserProvider{
				UserID:        existingUser.ID,
				Provider:      "wechat_mini_program",
				ProviderUID:   *openID,
				WechatUnionID: unionID,
			}
			if err := s.userRepo.CreateUserProvider(ctx, &userProvider); err != nil { // Pass context
				logger.Error(ctx, "Failed to create user provider", // Use slog.ErrorContext
					"userId", existingUser.ID,
					"provider", "wechat_mini_program",
					"providerUID", *openID,
					"unionID", unionID,
					"error", err)
				return 0, fmt.Errorf("failed to create user provider: %w", err)
			}
			logger.Info(ctx, "UserProvider created successfully", // Use slog.InfoContext
				"userId", existingUser.ID,
				"provider", "wechat_mini_program",
				"providerUID", *openID,
				"unionID", unionID)
			return existingUser.ID, nil
		} else if err != gorm.ErrRecordNotFound {
			logger.Error(ctx, "Failed to find user by unionID", "unionID", *unionID, "error", err) // Use slog.ErrorContext
			return 0, fmt.Errorf("database error: %w", err)
		}
	}

	// Convert DTO to User model.
	user := req.ToModel()

	// Call the repository layer to create the user.
	err := s.userRepo.CreateUser(ctx, user) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to create user", // Use slog.ErrorContext
			"name", user.Name,
			"openId", openID,
			"error", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// Create UserProvider record.
	userProvider := models.UserProvider{
		UserID:      user.ID,
		Provider:    "wechat_mini_program",
		ProviderUID: *openID,
	}
	// If unionID exists, set it.
	if unionID != nil && *unionID != "" {
		userProvider.WechatUnionID = unionID
	}

	// Call the repository layer to create UserProvider.
	err = s.userRepo.CreateUserProvider(ctx, &userProvider) // Pass context
	if err != nil {
		logger.Error(ctx, "Failed to create user provider", // Use slog.ErrorContext
			"userId", user.ID,
			"provider", "wechat_mini_program",
			"providerUID", *openID,
			"unionID", unionID,
			"error", err)
		return 0, fmt.Errorf("failed to create user provider: %w", err)
	}

	logger.Info(ctx, "User and UserProvider created successfully", // Use slog.InfoContext
		"userId", user.ID,
		"provider", "wechat_mini_program",
		"providerUID", *openID,
		"unionID", unionID)

	return user.ID, nil
}

// LoginFromWechatMiniProgram logs in a user via WeChat Mini Program.
func (s *userService) LoginFromWechatMiniProgram(ctx context.Context, unionID *string, openID *string) (string, string, error) {
	// If both unionID and openID are nil, return an error directly.
	if unionID == nil && openID == nil {
		return "", "", errors.ErrUserNotFound
	}

	var user *models.User
	var err error

	// Priority strategy: unionID > openID
	// Try finding the user by unionID first (if provided).
	if unionID != nil && *unionID != "" {
		user, err = s.userRepo.GetUserByUnionID(ctx, *unionID) // Pass context
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error(ctx, "Failed to find user by unionID", "unionID", *unionID, "error", err) // Use slog.ErrorContext
			return "", "", fmt.Errorf("database error: %w", err)
		}
	}

	// If not found by unionID, try finding by openID.
	if user == nil && openID != nil && *openID != "" {
		user, err = s.userRepo.GetUserByProvider(ctx, "wechat_mini_program", *openID) // Pass context
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error(ctx, "Failed to find user by provider", "provider", "wechat_mini_program", "providerUID", *openID, "error", err) // Use slog.ErrorContext
			return "", "", fmt.Errorf("database error: %w", err)
		}
	}

	// If still not found, return user not found error.
	if user == nil || user.ID == 0 {
		return "", "", errors.ErrUserNotFound
	}

	// Check if the user is banned.
	if user.IsBanned {
		return "", "", errors.ErrUserBanned
	}

	// Generate JWT access token.
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate access token", "error", err, "userId", user.ID) // Use slog.ErrorContext
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate JWT refresh token.
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate refresh token", "error", err, "userId", user.ID) // Use slog.ErrorContext
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login time.
	now := time.Now()
	s.userRepo.UpdateUser(ctx, user.ID, map[string]interface{}{"last_login": now}) // Pass context

	logger.Info(ctx, "WeChat mini program login successful", // Use slog.InfoContext
		"userId", user.ID,
		"unionID", unionID,
		"openID", openID)

	return accessToken, refreshToken, nil
}

/*
WeChat OAuth2.0 (App/Web)
*/

// WechatOAuthTokenResponse is the response structure for WeChat OAuth2 code exchange.
// [Reference] WeChat Login for Mobile Apps: https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html
// [Reference] WeChat Login for Web Apps: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
type WechatOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid,omitempty"`
	ErrCode      int    `json:"errcode,omitempty"`
	ErrMsg       string `json:"errmsg,omitempty"`
}

// ExchangeWechatOAuth handles WeChat OAuth2.0 (automatically determines login/registration).
func (s *userService) ExchangeWechatOAuth(ctx context.Context, req *dto.WechatOAuthRequest) (string, string, bool, error) {
	// 1. Exchange code for access_token and openid/unionid.
	appid := ""
	secret := ""
	if req.ClientType == "web" {
		appid = s.config.Wechat.Web.AppID
		secret = s.config.Wechat.Web.Secret
	} else if req.ClientType == "app" {
		appid = s.config.Wechat.App.AppID
		secret = s.config.Wechat.App.Secret
	} else {
		return "", "", false, fmt.Errorf("invalid client_type")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appid, secret, req.Code)
	client := &http.Client{Timeout: 10 * time.Second}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil) // Use http.NewRequestWithContext
	if err != nil {
		logger.Error(ctx, "Failed to create http request for wechat oauth", "error", err, "url", url)
		return "", "", false, fmt.Errorf("failed to create http request for wechat oauth: %w", err)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error(ctx, "Failed to request wechat oauth", "error", err, "url", url)
		return "", "", false, fmt.Errorf("failed to request wechat oauth: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp WechatOAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		logger.Error(ctx, "Failed to decode wechat oauth response", "error", err)
		return "", "", false, fmt.Errorf("failed to decode wechat oauth response: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		logger.Error(ctx, "Wechat oauth error", "errcode", tokenResp.ErrCode, "errmsg", tokenResp.ErrMsg)
		return "", "", false, fmt.Errorf("wechat oauth error: %s (%d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	openid := tokenResp.OpenID
	unionid := tokenResp.UnionID

	// 2. Check if openid/unionid is already bound to a user.
	var user *models.User
	isNewUser := false

	// Case 1: User found directly by openid.
	if openid != "" {
		user, err = s.userRepo.GetUserByProvider(ctx, "wechat", openid) // Pass context
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error(ctx, "Failed to find user by openid", "openid", openid, "error", err) // Use slog.ErrorContext
			return "", "", false, fmt.Errorf("failed to find user by openid: %w", err)
		}
	}

	// Case 2: If not found by openid and unionid is not empty, try finding by unionid.
	if (user == nil || user.ID == 0) && unionid != "" {
		user, err = s.userRepo.GetUserByUnionID(ctx, unionid) // Pass context
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error(ctx, "Failed to find user by unionid", "unionid", unionid, "error", err) // Use slog.ErrorContext
			return "", "", false, fmt.Errorf("failed to find user by unionid: %w", err)
		}

		// Create UserProvider record if an existing user is found by unionid.
		if user != nil && user.ID > 0 {
			userProvider := models.UserProvider{
				UserID:        user.ID,
				Provider:      "wechat",
				ProviderUID:   openid,
				WechatUnionID: &unionid, // unionid might be empty, so use a pointer.
			}
			if err := s.userRepo.CreateUserProvider(ctx, &userProvider); err != nil { // Pass context
				logger.Error(ctx, "Failed to create user provider for existing user by unionid", "error", err) // Use slog.ErrorContext
				return "", "", false, fmt.Errorf("failed to create user provider: %w", err)
			}
			logger.Info(ctx, "UserProvider created successfully for unionid", // Use slog.InfoContext
				"userId", user.ID,
				"provider", "wechat",
				"providerUID", openid,
				"unionID", unionid)
		}
	}

	// Case 3: User not found by either unionid or openid; create a new user.
	if user == nil || user.ID == 0 {
		// Create a new user.
		user = &models.User{
			Name:   fmt.Sprintf("微信用户_%s", openid[len(openid)-6:]), // WeChat User_xxxxxx
			Locale: "zh",                                           // Default to Chinese
		}
		if err := s.userRepo.CreateUser(ctx, user); err != nil { // Pass context
			logger.Error(ctx, "Failed to create new user from wechat oauth", "error", err) // Use slog.ErrorContext
			return "", "", false, fmt.Errorf("failed to create user: %w", err)
		}
		userProvider := models.UserProvider{
			UserID:      user.ID,
			Provider:    "wechat",
			ProviderUID: openid,
		}
		if unionid != "" {
			userProvider.WechatUnionID = &unionid
		}
		if err := s.userRepo.CreateUserProvider(ctx, &userProvider); err != nil { // Pass context
			logger.Error(ctx, "Failed to create user provider for new user from wechat oauth", "error", err) // Use slog.ErrorContext
			return "", "", false, fmt.Errorf("failed to create user provider: %w", err)
		}
		isNewUser = true

		logger.Info(ctx, "New user created from WeChat OAuth", // Use slog.InfoContext
			"userId", user.ID,
			"provider", "wechat",
			"providerUID", openid,
			"unionID", unionid)
	}

	// Check if the user is banned.
	if user.IsBanned {
		return "", "", false, errors.ErrUserBanned
	}

	// 3. Generate JWT access token.
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate access token from wechat oauth", "error", err, "userID", user.ID) // Use slog.ErrorContext
		return "", "", false, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 4. Generate JWT refresh token.
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate refresh token from wechat oauth", "error", err, "userID", user.ID) // Use slog.ErrorContext
		return "", "", false, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login time.
	now := time.Now()
	s.userRepo.UpdateUser(ctx, user.ID, map[string]interface{}{"last_login": now}) // Pass context

	return accessToken, refreshToken, isNewUser, nil
}

// BindWechatAccount binds a WeChat account to an existing user.
func (s *userService) BindWechatAccount(ctx context.Context, userID uint, req *dto.BindWechatAccountRequest, authenticatedUser *models.User) error {
	// Permission check: ensure the user can only bind their own account.
	if userID != authenticatedUser.ID {
		logger.Warn(ctx, "Permission denied for WeChat account binding", "userId", userID, "requesterId", authenticatedUser.ID) // Use slog.WarnContext
		return errors.ErrPermissionDenied
	}

	// 1. Exchange code for access_token and openid/unionid.
	appid := ""
	secret := ""
	if req.ClientType == "web" {
		appid = s.config.Wechat.Web.AppID
		secret = s.config.Wechat.Web.Secret
	} else if req.ClientType == "app" {
		appid = s.config.Wechat.App.AppID
		secret = s.config.Wechat.App.Secret
	} else {
		return fmt.Errorf("invalid client_type")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appid, secret, req.Code)
	client := &http.Client{Timeout: 10 * time.Second}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil) // Use http.NewRequestWithContext
	if err != nil {
		logger.Error(ctx, "Failed to create http request for wechat oauth binding", "error", err, "url", url)
		return fmt.Errorf("failed to create http request for wechat oauth binding: %w", err)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error(ctx, "Failed to request wechat oauth for binding", "error", err, "url", url)
		return fmt.Errorf("failed to request wechat oauth: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp WechatOAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		logger.Error(ctx, "Failed to decode wechat oauth response for binding", "error", err)
		return fmt.Errorf("failed to decode wechat oauth response: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		logger.Error(ctx, "Wechat oauth error for binding", "errcode", tokenResp.ErrCode, "errmsg", tokenResp.ErrMsg)
		return fmt.Errorf("wechat oauth error: %s (%d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	openid := tokenResp.OpenID
	unionid := tokenResp.UnionID

	// 2. Check if this WeChat account is already bound to another user.
	existingProvider, err := s.userRepo.GetUserByProvider(ctx, "wechat", openid) // Pass context
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error(ctx, "Failed to check existing WeChat provider for binding", "openId", openid, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check existing provider: %w", err)
	}

	if existingProvider != nil && existingProvider.ID != 0 {
		return errors.ErrProviderAlreadyBound
	}

	// 3. Check if the current user has already bound a WeChat account.
	_, err = s.userRepo.GetUserProvider(ctx, userID, "wechat") // Pass context
	if err == nil {
		return errors.ErrProviderAlreadyBound
	} else if err != gorm.ErrRecordNotFound {
		logger.Error(ctx, "Failed to check user's WeChat provider for binding", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// 4. Create the binding record.
	userProvider := &models.UserProvider{
		UserID:      userID,
		Provider:    "wechat",
		ProviderUID: openid,
	}
	if unionid != "" {
		userProvider.WechatUnionID = &unionid
	}

	if err := s.userRepo.CreateUserProvider(ctx, userProvider); err != nil { // Pass context
		logger.Error(ctx, "Failed to bind WeChat account", // Use slog.ErrorContext
			"userId", userID,
			"provider", "wechat",
			"providerUID", openid,
			"unionID", unionid,
			"error", err)
		return fmt.Errorf("failed to bind WeChat account: %w", err)
	}

	logger.Info(ctx, "WeChat account bound successfully", // Use slog.InfoContext
		"userId", userID,
		"provider", "wechat",
		"providerUID", openid,
		"unionID", unionid)

	return nil
}

// UnbindWechatAccount unbinds a WeChat account from a user.
func (s *userService) UnbindWechatAccount(ctx context.Context, userID uint, authenticatedUser *models.User) error {
	// Permission check: ensure the user can only unbind their own account.
	if userID != authenticatedUser.ID {
		logger.Warn(ctx, "Permission denied for WeChat account unbinding", "userId", userID, "requesterId", authenticatedUser.ID) // Use slog.WarnContext
		return errors.ErrPermissionDenied
	}

	// Check if the user has a verified email, otherwise unbinding could lock them out.
	user, err := s.GetUser(ctx, userID) // Pass context
	if err != nil {
		return err
	}

	if user.Email == nil || !user.IsEmailVerified {
		return errors.ErrEmailNotVerified
	}

	// Check if a WeChat account is bound.
	_, err = s.userRepo.GetUserProvider(ctx, userID, "wechat") // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProviderNotBound
		}
		logger.Error(ctx, "Failed to check user's WeChat provider for unbinding", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// Delete the binding record.
	if err := s.userRepo.DeleteUserProvider(ctx, userID, "wechat"); err != nil { // Pass context
		logger.Error(ctx, "Failed to unbind WeChat account", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to unbind WeChat account: %w", err)
	}

	logger.Info(ctx, "WeChat account unbound successfully", "userId", userID) // Use slog.InfoContext
	return nil
}

/*
Google OAuth2.0 (App/Web)
*/

// ExchangeGoogleOAuth handles Google OAuth2.0 (automatically determines login/registration).
func (s *userService) ExchangeGoogleOAuth(ctx context.Context, req *dto.GoogleOAuthRequest) (string, string, bool, error) {
	// 1. Exchange auth code for user information.
	googleUserInfo, err := s.googleOAuthService.ExchangeCodeForUserInfo(
		ctx, // Pass context
		req.Code,
		req.CodeVerifier,
		req.RedirectURI,
		req.ClientType,
	)
	if err != nil {
		logger.Error(ctx, "Failed to exchange Google OAuth code", "error", err) // Use slog.ErrorContext
		return "", "", false, err
	}

	// 2. Check if the user has already registered via Google.
	var user *models.User
	isNewUser := false

	// Case 1: User already exists.
	user, err = s.userRepo.GetUserByProvider(ctx, "google", googleUserInfo.ID) // Pass context
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error(ctx, "Failed to find user by Google provider", // Use slog.ErrorContext
			"providerUID", googleUserInfo.ID,
			"error", err)
		return "", "", false, fmt.Errorf("failed to find user: %w", err)
	}

	// Case 2: User does not exist, register a new user.
	if user == nil || user.ID == 0 {
		// Check if the email is already used by another user.
		if _, err := s.userRepo.GetUserByField(ctx, "email", googleUserInfo.Email, true); err == nil { // Pass context
			return "", "", false, errors.ErrEmailAlreadyExists
		}

		// Create new User model.
		user = &models.User{
			Name:            googleUserInfo.Name,
			Email:           &googleUserInfo.Email,
			Locale:          "en", // Default language
			IsEmailVerified: true, // Email from Google is considered verified.
		}

		// Use avatar URL from Google if provided.
		if googleUserInfo.Picture != "" {
			user.AvatarURL = &googleUserInfo.Picture
		}

		// Use language preference from Google if provided.
		if googleUserInfo.Locale != "" {
			user.Locale = googleUserInfo.Locale
		}

		// Call repository layer to create the user.
		if err := s.userRepo.CreateUser(ctx, user); err != nil { // Pass context
			logger.Error(ctx, "Failed to create user from Google registration", // Use slog.ErrorContext
				"email", googleUserInfo.Email,
				"googleId", googleUserInfo.ID,
				"error", err)
			return "", "", false, fmt.Errorf("failed to create user: %w", err)
		}

		// Create UserProvider record.
		userProvider := models.UserProvider{
			UserID:      user.ID,
			Provider:    "google",
			ProviderUID: googleUserInfo.ID,
		}

		if err := s.userRepo.CreateUserProvider(ctx, &userProvider); err != nil { // Pass context
			logger.Error(ctx, "Failed to create user provider for Google registration", // Use slog.ErrorContext
				"userId", user.ID,
				"provider", "google",
				"providerUID", googleUserInfo.ID,
				"error", err)
			return "", "", false, fmt.Errorf("failed to create user provider: %w", err)
		}
		isNewUser = true

		logger.Info(ctx, "User registered successfully with Google", // Use slog.InfoContext
			"userId", user.ID,
			"email", googleUserInfo.Email,
			"providerUID", googleUserInfo.ID)
	}

	// Check if the user is banned.
	if user.IsBanned {
		logger.Warn(ctx, "User is banned", "userId", user.ID, "email", googleUserInfo.Email) // Use slog.WarnContext
		return "", "", false, errors.ErrUserBanned
	}

	// 3. Generate JWT access token.
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate access token from google oauth", "error", err, "userId", user.ID) // Use slog.ErrorContext
		return "", "", false, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 4. Generate JWT refresh token.
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		logger.Error(ctx, "Failed to generate refresh token from google oauth", "error", err, "userId", user.ID) // Use slog.ErrorContext
		return "", "", false, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login time.
	now := time.Now()
	s.userRepo.UpdateUser(ctx, user.ID, map[string]interface{}{"last_login": now}) // Pass context

	return accessToken, refreshToken, isNewUser, nil
}

// BindGoogleAccount binds a Google account to an existing user.
func (s *userService) BindGoogleAccount(ctx context.Context, userID uint, req *dto.BindGoogleAccountRequest, authenticatedUser *models.User) error {
	// Permission check: ensure the user can only bind their own account.
	if userID != authenticatedUser.ID {
		logger.Warn(ctx, "Permission denied for Google account binding", "userId", userID, "requesterId", authenticatedUser.ID) // Use slog.WarnContext
		return errors.ErrPermissionDenied
	}

	// 1. Exchange auth code for user information.
	googleUserInfo, err := s.googleOAuthService.ExchangeCodeForUserInfo(
		ctx, // Pass context
		req.Code,
		req.CodeVerifier,
		req.RedirectURI,
		req.ClientType,
	)
	if err != nil {
		logger.Error(ctx, "Failed to exchange Google OAuth code for binding", "error", err, "userId", userID) // Use slog.ErrorContext
		return err
	}

	// 2. Check if this Google account is already bound to another user.
	existingProvider, err := s.userRepo.GetUserByProvider(ctx, "google", googleUserInfo.ID) // Pass context
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error(ctx, "Failed to check existing Google provider for binding", "googleId", googleUserInfo.ID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check existing provider: %w", err)
	}

	if existingProvider != nil && existingProvider.ID != 0 {
		return errors.ErrProviderAlreadyBound
	}

	// 3. Check if the current user has already bound a Google account.
	_, err = s.userRepo.GetUserProvider(ctx, userID, "google") // Pass context
	if err == nil {
		return errors.ErrProviderAlreadyBound
	} else if err != gorm.ErrRecordNotFound {
		logger.Error(ctx, "Failed to check user's Google provider for binding", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// 4. Create the binding record.
	userProvider := &models.UserProvider{
		UserID:      userID,
		Provider:    "google",
		ProviderUID: googleUserInfo.ID,
	}

	if err := s.userRepo.CreateUserProvider(ctx, userProvider); err != nil { // Pass context
		logger.Error(ctx, "Failed to bind Google account", // Use slog.ErrorContext
			"userId", userID,
			"provider", "google",
			"providerUID", googleUserInfo.ID,
			"error", err)
		return fmt.Errorf("failed to bind Google account: %w", err)
	}

	logger.Info(ctx, "Google account bound successfully", // Use slog.InfoContext
		"userId", userID,
		"provider", "google",
		"providerUID", googleUserInfo.ID)

	return nil
}

// UnbindGoogleAccount unbinds a Google account from a user.
func (s *userService) UnbindGoogleAccount(ctx context.Context, userID uint, authenticatedUser *models.User) error {
	// Permission check: ensure the user can only unbind their own account.
	if userID != authenticatedUser.ID {
		logger.Warn(ctx, "Permission denied for Google account unbinding", "userId", userID, "requesterId", authenticatedUser.ID) // Use slog.WarnContext
		return errors.ErrPermissionDenied
	}

	// Check if a Google account is bound.
	_, err := s.userRepo.GetUserProvider(ctx, userID, "google") // Pass context
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProviderNotBound
		}
		logger.Error(ctx, "Failed to check user's Google provider for unbinding", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// Delete the binding record.
	if err := s.userRepo.DeleteUserProvider(ctx, userID, "google"); err != nil { // Pass context
		logger.Error(ctx, "Failed to unbind Google account", "userId", userID, "error", err) // Use slog.ErrorContext
		return fmt.Errorf("failed to unbind Google account: %w", err)
	}

	logger.Info(ctx, "Google account unbound successfully", "userId", userID) // Use slog.InfoContext
	return nil
}
