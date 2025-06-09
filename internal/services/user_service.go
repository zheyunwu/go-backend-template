package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/jwt"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 定义用户相关的业务逻辑接口
type UserService interface {
	/* 面向Admin的业务逻辑 */
	ListUsers(params *query_params.QueryParams, includeSoftDeleted ...bool) ([]models.User, *response.Pagination, error)
	GetUser(id uint, includeSoftDeleted ...bool) (*models.User, error)
	CreateUser(req *dto.RegisterWithPasswordRequest) (uint, error)
	UpdateUser(id uint, req *dto.UpdateProfileRequest) error
	DeleteUser(id uint) error
	RestoreUser(id uint) error          // 软删除的用户恢复
	BanUser(id uint, banned bool) error // banned为true时封禁，false时解除封禁

	/* Auth逻辑 */
	CheckUserExists(fieldType string, value string) (bool, error)
	UpdatePassword(userID uint, currentPassword, newPassword string) error

	/* 传统注册登录相关 */
	RegisterWithPassword(req *dto.RegisterWithPasswordRequest) (uint, error)
	LoginWithPassword(emailOrPhone, password string) (string, error)

	/* 邮箱验证相关 */
	SendEmailVerification(email string) error
	VerifyEmail(email, code string) error

	/* 密码重置相关 */
	SendPasswordReset(email string) error
	ResetPassword(email, resetToken, newPassword string) error

	/* 微信小程序端 */
	RegisterFromWechatMiniProgram(req *dto.RegisterFromWechatMiniProgramRequest, unionID *string, openID *string) (uint, error)
	LoginFromWechatMiniProgram(unionID *string, openID *string) (string, error)

	/* 微信OAuth2.0（App/Web端） */
	ExchangeWechatOAuth(req *dto.WechatOAuthRequest) (string, bool, error)
	BindWechatAccount(userID uint, req *dto.BindWechatAccountRequest, authenticatedUser *models.User) error
	UnbindWechatAccount(userID uint, authenticatedUser *models.User) error
	/* Google OAuth2.0（App/Web端） */
	ExchangeGoogleOAuth(req *dto.GoogleOAuthRequest) (string, bool, error)
	BindGoogleAccount(userID uint, req *dto.BindGoogleAccountRequest, authenticatedUser *models.User) error
	UnbindGoogleAccount(userID uint, authenticatedUser *models.User) error
}

// userService 用户服务实现
type userService struct {
	config              *config.Config
	userRepo            repositories.UserRepository
	googleOAuthService  GoogleOAuthService
	emailService        EmailService
	verificationService VerificationService
}

// NewUserService 创建一个用户服务实例
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
面向Admin的业务逻辑
*/

// ListUsers 获取用户列表
func (s *userService) ListUsers(params *query_params.QueryParams, includeSoftDeleted ...bool) ([]models.User, *response.Pagination, error) {
	// 调用Repo层 获取用户列表
	userList, total, err := s.userRepo.ListUsers(params, includeSoftDeleted...)
	if err != nil {
		slog.Error("Failed to list users", "error", err)
		return nil, nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 没有数据时返回空数组
	if len(userList) == 0 {
		userList = []models.User{}
	}

	// 构建分页信息
	pagination := &response.Pagination{
		TotalCount:  int(total),
		PageSize:    params.Limit,
		CurrentPage: params.Page,
		TotalPages:  (int(total) + params.Limit - 1) / params.Limit,
	}

	return userList, pagination, nil
}

// GetUser 获取用户详情
func (s *userService) GetUser(id uint, includeSoftDeleted ...bool) (*models.User, error) {
	// 调用repo层获取用户
	user, err := s.userRepo.GetUser(id, includeSoftDeleted...)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		slog.Error("Failed to get user from repository", "userId", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser 创建用户（传统Email/Phone + 密码）
func (s *userService) CreateUser(req *dto.RegisterWithPasswordRequest) (uint, error) {
	// 验证至少提供email或phone之一
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return 0, errors.ErrEmailOrPhoneNotProvided
	}

	// 验证是否有重复的邮箱或手机号
	if req.Email != nil && *req.Email != "" {
		if _, err := s.userRepo.GetUserByField("email", *req.Email, true); err == nil {
			return 0, errors.ErrEmailAlreadyExists
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if _, err := s.userRepo.GetUserByField("phone", *req.Phone, true); err == nil {
			return 0, errors.ErrPhoneAlreadyExists
		}
	}

	// 对密码进行哈希处理
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return 0, fmt.Errorf("error processing password: %w", err)
	}
	// 使用DTO转换为User模型
	user := req.ToModel(hashedPassword)

	// 调用repo层创建用户
	if err := s.userRepo.CreateUser(user); err != nil {
		slog.Error("Failed to create user",
			"name", user.Name,
			"email", req.Email,
			"phone", req.Phone,
			"error", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// 如果用户提供了邮箱，自动发送验证邮件
	if req.Email != nil && *req.Email != "" {
		// 生成验证码
		code, err := s.verificationService.GenerateEmailVerificationCode(*req.Email)
		if err != nil {
			slog.Warn("Failed to generate email verification code after user creation", "userId", user.ID, "email", *req.Email, "error", err)
			// 这里不返回错误，因为用户已经创建成功，只是验证邮件发送失败
		} else {
			// 发送验证邮件，使用用户的语言偏好
			err = s.emailService.SendEmailVerification(*req.Email, user.Name, code, user.Locale)
			if err != nil {
				slog.Warn("Failed to send verification email after user creation", "userId", user.ID, "email", *req.Email, "error", err)
				// 这里也不返回错误，因为用户已经创建成功
			} else {
				slog.Info("Verification email sent successfully after user creation", "userId", user.ID, "email", *req.Email, "language", user.Locale)
			}
		}
	}

	return user.ID, nil
}

// hashPassword 使用bcrypt哈希密码
func hashPassword(password string) (string, error) {
	// 使用推荐的cost值(10-12)生成哈希密码
	// bcrypt自动添加随机盐值并将其包含在哈希结果中
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(id uint, req *dto.UpdateProfileRequest) error {
	// 检查用户是否存在
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// 将DTO转换为更新字段Map
	updates := req.ToUpdatesMap()

	// 如果更新了Email，检查是否已存在、并设定为未验证状态
	if req.Email != nil && *req.Email != "" && (user.Email == nil || *user.Email != *req.Email) {
		existingUser, err := s.userRepo.GetUserByField("email", *req.Email)
		if err != nil && err != gorm.ErrRecordNotFound {
			slog.Error("Failed to check existing email", "email", *req.Email, "error", err)
			return fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingUser != nil && existingUser.ID != 0 && existingUser.ID != id {
			slog.Warn("Email already exists", "email", *req.Email)
			return errors.ErrEmailAlreadyExists
		}
		// 如果Email已更新，设置为未验证状态
		updates["is_email_verified"] = false
	}

	// 调用repo层进行更新
	if err := s.userRepo.UpdateUser(id, updates); err != nil {
		slog.Error("Failed to update user", "userId", id, "error", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	// 检查用户是否存在
	_, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// 调用repo层删除用户
	if err := s.userRepo.DeleteUser(id); err != nil {
		slog.Error("Failed to delete user", "userId", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// RestoreUser 恢复软删除的用户
func (s *userService) RestoreUser(id uint) error {
	// 检查用户是否存在
	user, err := s.GetUser(id, true) // 包含软删除的记录
	if err != nil {
		return err
	}

	// 如果用户没有被软删除，直接返回
	if !user.DeletedAt.Valid {
		return nil
	}

	// 构建更新字段映射
	updates := map[string]interface{}{"deleted_at": nil}

	// 调用repo层恢复用户
	if err := s.userRepo.UpdateUser(id, updates, true); err != nil {
		slog.Error("Failed to restore user", "userId", id, "error", err)
		return fmt.Errorf("failed to restore user: %w", err)
	}

	return nil
}

// BanUser 封禁或解除封禁用户
func (s *userService) BanUser(id uint, isBanned bool) error {
	// 检查用户是否存在
	_, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// 构建更新字段映射
	updates := map[string]interface{}{
		"is_banned": isBanned,
	}
	if err := s.userRepo.UpdateUser(id, updates); err != nil {
		slog.Error("Failed to update ban status", "userId", id, "error", err)
		return fmt.Errorf("failed to update ban status: %w", err)
	}
	return nil
}

/*
Auth逻辑
*/

// CheckUserExists 检查用户是否存在
func (s *userService) CheckUserExists(fieldType string, value string) (bool, error) {
	var err error

	switch fieldType {
	case "mini_program_open_id":
		// 通过UserProvider表查询微信小程序openID
		_, err = s.userRepo.GetUserByProvider("wechat_mini_program", value)
	case "union_id":
		// 通过UserProvider表查询微信UnionID
		_, err = s.userRepo.GetUserByUnionID(value)
	default:
		// 其他字段直接在User表中查询
		_, err = s.userRepo.GetUserByField(fieldType, value)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdatePassword 更新用户密码
func (s *userService) UpdatePassword(userID uint, currentPassword, newPassword string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		slog.Error("Failed to find user", "userId", userID, "error", err)
		return fmt.Errorf("database error: %w", err)
	}

	// 验证当前密码
	if user.Password == nil || *user.Password == "" {
		slog.Warn("User has no password set", "userId", user.ID)
		return errors.ErrInvalidPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword))
	if err != nil {
		slog.Warn("Current password verification failed", "userId", user.ID)
		return errors.ErrInvalidPassword
	}

	// 哈希新密码
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		slog.Error("Failed to hash new password", "userId", user.ID, "error", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新用户密码
	updates := map[string]interface{}{
		"password": hashedPassword,
	}

	// 调用repo层更新用户密码
	if err := s.userRepo.UpdateUser(user.ID, updates); err != nil {
		slog.Error("Failed to update user password", "userId", user.ID, "error", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	slog.Info("Password updated successfully", "userId", user.ID)
	return nil
}

/*
传统注册登录相关
*/

// RegisterWithPassword 使用密码注册用户
func (s *userService) RegisterWithPassword(req *dto.RegisterWithPasswordRequest) (uint, error) {
	// 转交给CreateUser处理
	return s.CreateUser(req)
}

// LoginWithPassword 验证用户密码并生成JWT token
func (s *userService) LoginWithPassword(emailOrPhone, password string) (string, error) {
	var user *models.User
	var err error

	// 先尝试用邮箱查找用户
	user, err = s.userRepo.GetUserByField("email", emailOrPhone)
	if err != nil && err == gorm.ErrRecordNotFound {
		// 如果邮箱找不到，尝试用手机号查找
		user, err = s.userRepo.GetUserByField("phone", emailOrPhone)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return "", errors.ErrUserNotFound
			}
			slog.Error("Failed to find user", "emailOrPhone", emailOrPhone, "error", err)
			return "", fmt.Errorf("database error: %w", err)
		}
	} else if err != nil {
		slog.Error("Failed to find user", "emailOrPhone", emailOrPhone, "error", err)
		return "", fmt.Errorf("database error: %w", err)
	}

	// 检查用户是否被封禁
	if user.IsBanned {
		return "", errors.ErrUserBanned
	}

	// 检查有无密码
	if user.Password == nil || *user.Password == "" {
		slog.Warn("User has no password set", "userId", user.ID)
		return "", errors.ErrInvalidPassword
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		slog.Warn("Password verification failed", "userId", user.ID)
		return "", errors.ErrInvalidPassword
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		slog.Error("Failed to generate JWT", "error", err, "userId", user.ID)
		return "", fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.userRepo.UpdateUser(user.ID, map[string]interface{}{"last_login": now})

	return token, nil
}

/*
邮箱验证相关
*/

// SendEmailVerification 发送邮箱验证码
func (s *userService) SendEmailVerification(email string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByField("email", email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		slog.Error("Failed to find user by email", "email", email, "error", err)
		return fmt.Errorf("database error: %w", err)
	}

	// 检查邮箱是否已验证
	if user.IsEmailVerified {
		return errors.ErrEmailAlreadyVerified
	}

	// 生成验证码
	code, err := s.verificationService.GenerateEmailVerificationCode(email)
	if err != nil {
		slog.Error("Failed to generate email verification code", "email", email, "error", err)
		return fmt.Errorf("failed to generate verification code: %w", err)
	}

	// 发送验证邮件
	err = s.emailService.SendEmailVerification(email, user.Name, code, user.Locale)
	if err != nil {
		slog.Error("Failed to send verification email", "email", email, "error", err)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	slog.Info("Email verification sent successfully", "email", email)
	return nil
}

// VerifyEmail 验证邮箱
func (s *userService) VerifyEmail(email, code string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByField("email", email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		slog.Error("Failed to find user by email", "email", email, "error", err)
		return fmt.Errorf("database error: %w", err)
	}

	// 检查邮箱是否已验证
	if user.IsEmailVerified {
		return errors.ErrEmailAlreadyVerified
	}

	// 验证验证码
	isValid, err := s.verificationService.VerifyEmailVerificationCode(email, code)
	if err != nil {
		slog.Error("Failed to verify email verification code", "email", email, "error", err)
		return fmt.Errorf("failed to verify code: %w", err)
	}

	if !isValid {
		return errors.ErrInvalidVerificationCode
	}

	// 更新用户邮箱验证状态
	updates := map[string]interface{}{
		"is_email_verified": true,
	}

	if err := s.userRepo.UpdateUser(user.ID, updates); err != nil {
		slog.Error("Failed to update email verification status", "userId", user.ID, "error", err)
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	slog.Info("Email verified successfully", "email", email, "userId", user.ID)
	return nil
}

/*
密码重置相关
*/

// SendPasswordReset 发送密码重置邮件
func (s *userService) SendPasswordReset(email string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByField("email", email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		slog.Error("Failed to find user by email", "email", email, "error", err)
		return fmt.Errorf("database error: %w", err)
	}

	// 生成重置令牌
	token, err := s.verificationService.GeneratePasswordResetToken(email)
	if err != nil {
		slog.Error("Failed to generate password reset token", "email", email, "error", err)
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// 发送重置邮件
	err = s.emailService.SendPasswordReset(email, user.Name, token, user.Locale)
	if err != nil {
		slog.Error("Failed to send password reset email", "email", email, "error", err)
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	slog.Info("Password reset email sent successfully", "email", email)
	return nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(email, resetToken, newPassword string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByField("email", email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		slog.Error("Failed to find user by email", "email", email, "error", err)
		return fmt.Errorf("database error: %w", err)
	}

	// 验证重置令牌
	isValid, err := s.verificationService.VerifyPasswordResetToken(email, resetToken)
	if err != nil {
		slog.Error("Failed to verify password reset token", "email", email, "error", err)
		return fmt.Errorf("failed to verify reset token: %w", err)
	}

	if !isValid {
		return errors.ErrInvalidVerificationCode
	}

	// 哈希新密码
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		slog.Error("Failed to hash new password", "email", email, "error", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新用户密码
	updates := map[string]interface{}{
		"password": hashedPassword,
	}

	if err := s.userRepo.UpdateUser(user.ID, updates); err != nil {
		slog.Error("Failed to update user password", "userId", user.ID, "error", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	slog.Info("Password reset successfully", "email", email, "userId", user.ID)
	return nil
}

/*
微信小程序端
*/

// RegisterUserFromWechatMiniProgram 从微信小程序注册用户
func (s *userService) RegisterFromWechatMiniProgram(req *dto.RegisterFromWechatMiniProgramRequest, unionID *string, openID *string) (uint, error) {
	// 在生产环境下检查敏感内容
	// if openID != nil && *openID != "" {
	// 	// 检查用户昵称是否包含敏感内容
	// 	if err := utils.CheckSensitiveContent(req.Name, *openID, utils.SecuritySceneProfile); err != nil {
	// 		return 0, err
	// 	}
	// }

	// 验证openID是否提供
	if openID == nil || *openID == "" {
		return 0, errors.ErrUserNotFound // 或者定义一个新的错误类型
	}

	// 验证：用户是否已在微信小程序端注册过
	if _, err := s.userRepo.GetUserByProvider("wechat_mini_program", *openID); err == nil {
		return 0, errors.ErrUserAlreadyExists
	}

	// 若有UnionID，看一下用户是否已在APP端注册过，如有，则直接取其关联的UserID
	if unionID != nil && *unionID != "" {
		existingUser, err := s.userRepo.GetUserByUnionID(*unionID)
		if err == nil && existingUser != nil && existingUser.ID > 0 {
			// 用户已在APP端注册过，用这个用户的ID
			userProvider := models.UserProvider{
				UserID:        existingUser.ID,
				Provider:      "wechat_mini_program",
				ProviderUID:   *openID,
				WechatUnionID: unionID,
			}
			if err := s.userRepo.CreateUserProvider(&userProvider); err != nil {
				slog.Error("Failed to create user provider",
					"userId", existingUser.ID,
					"provider", "wechat_mini_program",
					"providerUID", *openID,
					"unionID", unionID,
					"error", err)
				return 0, fmt.Errorf("failed to create user provider: %w", err)
			}
			slog.Info("UserProvider created successfully",
				"userId", existingUser.ID,
				"provider", "wechat_mini_program",
				"providerUID", *openID,
				"unionID", unionID)
			return existingUser.ID, nil
		} else if err != gorm.ErrRecordNotFound {
			slog.Error("Failed to find user by unionID", "unionID", *unionID, "error", err)
			return 0, fmt.Errorf("database error: %w", err)
		}
	}

	// 验证是否有重复的邮箱或手机号
	if req.Email != nil && *req.Email != "" {
		if _, err := s.userRepo.GetUserByField("email", *req.Email, true); err == nil {
			return 0, errors.ErrEmailAlreadyExists
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		if _, err := s.userRepo.GetUserByField("phone", *req.Phone, true); err == nil {
			return 0, errors.ErrPhoneAlreadyExists
		}
	}

	// DTO转换为User模型
	user := req.ToModel()

	// 调用repo层创建用户
	err := s.userRepo.CreateUser(user)
	if err != nil {
		slog.Error("Failed to create user",
			"nickname", user.Name,
			"email", req.Email,
			"phone", req.Phone,
			"openId", openID,
			"error", err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	// 创建UserProvider记录
	userProvider := models.UserProvider{
		UserID:      user.ID,
		Provider:    "wechat_mini_program",
		ProviderUID: *openID,
	}
	// 如果有unionID，设置它
	if unionID != nil && *unionID != "" {
		userProvider.WechatUnionID = unionID
	}

	// 调用repo层创建UserProvider
	err = s.userRepo.CreateUserProvider(&userProvider)
	if err != nil {
		slog.Error("Failed to create user provider",
			"userId", user.ID,
			"provider", "wechat_mini_program",
			"providerUID", *openID,
			"unionID", unionID,
			"error", err)
		return 0, fmt.Errorf("failed to create user provider: %w", err)
	}

	slog.Info("User and UserProvider created successfully",
		"userId", user.ID,
		"provider", "wechat_mini_program",
		"providerUID", *openID,
		"unionID", unionID)

	return user.ID, nil
}

// LoginFromWechatMiniProgram 微信小程序登录
func (s *userService) LoginFromWechatMiniProgram(unionID *string, openID *string) (string, error) {
	// 如果unionID和openID都为nil，则直接返回错误
	if unionID == nil && openID == nil {
		return "", errors.ErrUserNotFound
	}

	var user *models.User
	var err error

	// 优先级策略：unionID > openID
	// 先尝试用unionID查找用户（如果提供了的话）
	if unionID != nil && *unionID != "" {
		user, err = s.userRepo.GetUserByUnionID(*unionID)
		if err != nil && err != gorm.ErrRecordNotFound {
			slog.Error("Failed to find user by unionID", "unionID", *unionID, "error", err)
			return "", fmt.Errorf("database error: %w", err)
		}
	}

	// 如果通过unionID没有找到用户，再尝试用openID查找
	if user == nil && openID != nil && *openID != "" {
		user, err = s.userRepo.GetUserByProvider("wechat_mini_program", *openID)
		if err != nil && err != gorm.ErrRecordNotFound {
			slog.Error("Failed to find user by provider", "provider", "wechat_mini_program", "providerUID", *openID, "error", err)
			return "", fmt.Errorf("database error: %w", err)
		}
	}

	// 如果还是没有找到用户，返回用户未找到错误
	if user == nil || user.ID == 0 {
		return "", errors.ErrUserNotFound
	}

	// 检查用户是否被封禁
	if user.IsBanned {
		return "", errors.ErrUserBanned
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		slog.Error("Failed to generate JWT", "error", err, "userId", user.ID)
		return "", fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.userRepo.UpdateUser(user.ID, map[string]interface{}{"last_login": now})

	slog.Info("WeChat mini program login successful",
		"userId", user.ID,
		"unionID", unionID,
		"openID", openID)

	return token, nil
}

/*
微信OAuth2.0 （App/Web端）
*/

// WechatOAuthTokenResponse 微信OAuth2 code换token响应
// [参考] APP端微信登录：https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html
// [参考] Web端微信登录：https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
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

// ExchangeWechatOAuth 处理微信OAuth2.0（自动判断登录/注册）
func (s *userService) ExchangeWechatOAuth(req *dto.WechatOAuthRequest) (string, bool, error) {
	// 1. 用 code 换取 access_token 和 openid/unionid
	appid := ""
	secret := ""
	if req.ClientType == "web" {
		appid = s.config.Wechat.Web.AppID
		secret = s.config.Wechat.Web.Secret
	} else if req.ClientType == "app" {
		appid = s.config.Wechat.App.AppID
		secret = s.config.Wechat.App.Secret
	} else {
		return "", false, fmt.Errorf("invalid client_type")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appid, secret, req.Code)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", false, fmt.Errorf("failed to request wechat oauth: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp WechatOAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", false, fmt.Errorf("failed to decode wechat oauth response: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return "", false, fmt.Errorf("wechat oauth error: %s (%d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	openid := tokenResp.OpenID
	unionid := tokenResp.UnionID

	// 2. 检查 openid/unionid 是否已绑定用户
	var user *models.User
	isNewUser := false

	// 情况1: openid直接找到用户
	if openid != "" {
		user, err = s.userRepo.GetUserByProvider("wechat", openid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return "", false, fmt.Errorf("failed to find user by openid: %w", err)
		}
	}

	// 情况2: unionid不为空时，尝试通过unionid查找用户
	if (user == nil || user.ID == 0) && unionid != "" {
		user, err = s.userRepo.GetUserByUnionID(unionid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return "", false, fmt.Errorf("failed to find user by unionid: %w", err)
		}

		// 创建 UserProvider 记录
		if user != nil && user.ID > 0 {
			userProvider := models.UserProvider{
				UserID:        user.ID,
				Provider:      "wechat",
				ProviderUID:   openid,
				WechatUnionID: &unionid, // unionid可能为空，所以用指针
			}
			if err := s.userRepo.CreateUserProvider(&userProvider); err != nil {
				return "", false, fmt.Errorf("failed to create user provider: %w", err)
			}
			slog.Info("UserProvider created successfully for unionid",
				"userId", user.ID,
				"provider", "wechat",
				"providerUID", openid,
				"unionID", unionid)
		}
	}

	// 情况3: unionid和openid都没有找到用户
	if user == nil || user.ID == 0 {
		// 创建新用户
		user = &models.User{
			Name:   fmt.Sprintf("微信用户_%s", openid[len(openid)-6:]),
			Locale: "zh",
		}
		if err := s.userRepo.CreateUser(user); err != nil {
			return "", false, fmt.Errorf("failed to create user: %w", err)
		}
		userProvider := models.UserProvider{
			UserID:      user.ID,
			Provider:    "wechat",
			ProviderUID: openid,
		}
		if unionid != "" {
			userProvider.WechatUnionID = &unionid
		}
		if err := s.userRepo.CreateUserProvider(&userProvider); err != nil {
			return "", false, fmt.Errorf("failed to create user provider: %w", err)
		}
		isNewUser = true

		slog.Info("New user created from WeChat OAuth",
			"userId", user.ID,
			"provider", "wechat",
			"providerUID", openid,
			"unionID", unionid)
	}

	// 检查用户是否被封禁
	if user.IsBanned {
		return "", false, errors.ErrUserBanned
	}

	// 3. 返回 JWT token
	token, err := jwt.GenerateToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		return "", false, fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.userRepo.UpdateUser(user.ID, map[string]interface{}{"last_login": now})

	return token, isNewUser, nil
}

// BindWechatAccount 绑定微信账号
func (s *userService) BindWechatAccount(userID uint, req *dto.BindWechatAccountRequest, authenticatedUser *models.User) error {
	// 权限检查：确保用户只能绑定自己的账号
	if userID != authenticatedUser.ID {
		slog.Warn("Permission denied for WeChat account binding", "userId", userID, "requesterId", authenticatedUser.ID)
		return errors.ErrPermissionDenied
	}

	// 1. 用 code 换取 access_token 和 openid/unionid
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
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to request wechat oauth: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp WechatOAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode wechat oauth response: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return fmt.Errorf("wechat oauth error: %s (%d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	openid := tokenResp.OpenID
	unionid := tokenResp.UnionID

	// 2. 检查该微信账号是否已被其他用户绑定
	existingProvider, err := s.userRepo.GetUserByProvider("wechat", openid)
	if err != nil && err != gorm.ErrRecordNotFound {
		slog.Error("Failed to check existing WeChat provider", "openId", openid, "error", err)
		return fmt.Errorf("failed to check existing provider: %w", err)
	}

	if existingProvider != nil && existingProvider.ID != 0 {
		return errors.ErrProviderAlreadyBound
	}

	// 3. 检查当前用户是否已绑定微信账号
	_, err = s.userRepo.GetUserProvider(userID, "wechat")
	if err == nil {
		return errors.ErrProviderAlreadyBound
	} else if err != gorm.ErrRecordNotFound {
		slog.Error("Failed to check user's WeChat provider", "userId", userID, "error", err)
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// 4. 创建绑定记录
	userProvider := &models.UserProvider{
		UserID:      userID,
		Provider:    "wechat",
		ProviderUID: openid,
	}
	if unionid != "" {
		userProvider.WechatUnionID = &unionid
	}

	if err := s.userRepo.CreateUserProvider(userProvider); err != nil {
		slog.Error("Failed to bind WeChat account",
			"userId", userID,
			"provider", "wechat",
			"providerUID", openid,
			"unionID", unionid,
			"error", err)
		return fmt.Errorf("failed to bind WeChat account: %w", err)
	}

	slog.Info("WeChat account bound successfully",
		"userId", userID,
		"provider", "wechat",
		"providerUID", openid,
		"unionID", unionid)

	return nil
}

// UnbindWechatAccount 解绑微信账号
func (s *userService) UnbindWechatAccount(userID uint, authenticatedUser *models.User) error {
	// 权限检查：确保用户只能解绑自己的账号
	if userID != authenticatedUser.ID {
		slog.Warn("Permission denied for WeChat account unbinding", "userId", userID, "requesterId", authenticatedUser.ID)
		return errors.ErrPermissionDenied
	}

	// 检查用户是否已提供验证过的邮箱，否则解绑后再也无法登录
	user, err := s.GetUser(userID)
	if err != nil {
		return err
	}

	if user.Email == nil || !user.IsEmailVerified {
		return errors.ErrEmailNotVerified
	}

	// 检查是否已绑定微信账号
	_, err = s.userRepo.GetUserProvider(userID, "wechat")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProviderNotBound
		}
		slog.Error("Failed to check user's WeChat provider", "userId", userID, "error", err)
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// 删除绑定记录
	if err := s.userRepo.DeleteUserProvider(userID, "wechat"); err != nil {
		slog.Error("Failed to unbind WeChat account", "userId", userID, "error", err)
		return fmt.Errorf("failed to unbind WeChat account: %w", err)
	}

	slog.Info("WeChat account unbound successfully", "userId", userID)
	return nil
}

/*
Google OAuth2.0（App/Web端）
*/

// ExchangeGoogleOAuth 处理Google OAuth2.0（自动判断登录/注册）
func (s *userService) ExchangeGoogleOAuth(req *dto.GoogleOAuthRequest) (string, bool, error) {
	// 1. 用auth code换取用户信息
	googleUserInfo, err := s.googleOAuthService.ExchangeCodeForUserInfo(
		context.Background(),
		req.Code,
		req.CodeVerifier,
		req.RedirectURI,
		req.ClientType,
	)
	if err != nil {
		slog.Error("Failed to exchange Google OAuth code", "error", err)
		return "", false, err
	}

	// 2. 检查用户是否已通过Google注册
	var user *models.User
	isNewUser := false

	// 情况1: 用户已存在
	user, err = s.userRepo.GetUserByProvider("google", googleUserInfo.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		slog.Error("Failed to find user by Google provider",
			"providerUID", googleUserInfo.ID,
			"error", err)
		return "", false, fmt.Errorf("failed to find user: %w", err)
	}

	// 情况2: 用户不存在，需要注册新用户
	if user == nil || user.ID == 0 {
		// 检查邮箱是否已被其他用户使用
		if _, err := s.userRepo.GetUserByField("email", googleUserInfo.Email, true); err == nil {
			return "", false, errors.ErrEmailAlreadyExists
		}

		// 创建新用户Model
		user = &models.User{
			Name:            googleUserInfo.Name,
			Email:           &googleUserInfo.Email,
			Locale:          "en", // 默认语言
			IsEmailVerified: true,
		}

		// 如果Google提供了头像URL，使用它
		if googleUserInfo.Picture != "" {
			user.AvatarURL = &googleUserInfo.Picture
		}

		// 如果Google提供了语言偏好，使用它
		if googleUserInfo.Locale != "" {
			user.Locale = googleUserInfo.Locale
		}

		// 调用repo层创建用户
		if err := s.userRepo.CreateUser(user); err != nil {
			slog.Error("Failed to create user from Google registration",
				"email", googleUserInfo.Email,
				"googleId", googleUserInfo.ID,
				"error", err)
			return "", false, fmt.Errorf("failed to create user: %w", err)
		}

		// 创建UserProvider记录
		userProvider := models.UserProvider{
			UserID:      user.ID,
			Provider:    "google",
			ProviderUID: googleUserInfo.ID,
		}

		if err := s.userRepo.CreateUserProvider(&userProvider); err != nil {
			slog.Error("Failed to create user provider for Google registration",
				"userId", user.ID,
				"provider", "google",
				"providerUID", googleUserInfo.ID,
				"error", err)
			return "", false, fmt.Errorf("failed to create user provider: %w", err)
		}
		isNewUser = true

		slog.Info("User registered successfully with Google",
			"userId", user.ID,
			"email", googleUserInfo.Email,
			"providerUID", googleUserInfo.ID)
	}

	// 检查用户是否被封禁
	if user.IsBanned {
		slog.Warn("User is banned", "userId", user.ID, "email", googleUserInfo.Email)
		return "", false, errors.ErrUserBanned
	}

	// 3. 返回JWT token
	token, err := jwt.GenerateToken(user.ID, user.Role, s.config.JWT.Secret, s.config.JWT.ExpireHours)
	if err != nil {
		slog.Error("Failed to generate JWT", "error", err, "userId", user.ID)
		return "", false, fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.userRepo.UpdateUser(user.ID, map[string]interface{}{"last_login": now})

	return token, isNewUser, nil
}

// BindGoogleAccount 绑定Google账号
func (s *userService) BindGoogleAccount(userID uint, req *dto.BindGoogleAccountRequest, authenticatedUser *models.User) error {
	// 权限检查：确保用户只能绑定自己的账号
	if userID != authenticatedUser.ID {
		slog.Warn("Permission denied for Google account binding", "userId", userID, "requesterId", authenticatedUser.ID)
		return errors.ErrPermissionDenied
	}

	// 1. 用auth code换取用户信息
	googleUserInfo, err := s.googleOAuthService.ExchangeCodeForUserInfo(
		context.Background(),
		req.Code,
		req.CodeVerifier,
		req.RedirectURI,
		req.ClientType,
	)
	if err != nil {
		slog.Error("Failed to exchange Google OAuth code for binding", "error", err, "userId", userID)
		return err
	}

	// 2. 检查该Google账号是否已被其他用户绑定
	existingProvider, err := s.userRepo.GetUserByProvider("google", googleUserInfo.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		slog.Error("Failed to check existing Google provider", "googleId", googleUserInfo.ID, "error", err)
		return fmt.Errorf("failed to check existing provider: %w", err)
	}

	if existingProvider != nil && existingProvider.ID != 0 {
		return errors.ErrProviderAlreadyBound
	}

	// 3. 检查当前用户是否已绑定Google账号
	_, err = s.userRepo.GetUserProvider(userID, "google")
	if err == nil {
		return errors.ErrProviderAlreadyBound
	} else if err != gorm.ErrRecordNotFound {
		slog.Error("Failed to check user's Google provider", "userId", userID, "error", err)
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// 4. 创建绑定记录
	userProvider := &models.UserProvider{
		UserID:      userID,
		Provider:    "google",
		ProviderUID: googleUserInfo.ID,
	}

	if err := s.userRepo.CreateUserProvider(userProvider); err != nil {
		slog.Error("Failed to bind Google account",
			"userId", userID,
			"provider", "google",
			"providerUID", googleUserInfo.ID,
			"error", err)
		return fmt.Errorf("failed to bind Google account: %w", err)
	}

	slog.Info("Google account bound successfully",
		"userId", userID,
		"provider", "google",
		"providerUID", googleUserInfo.ID)

	return nil
}

// UnbindGoogleAccount 解绑Google账号
func (s *userService) UnbindGoogleAccount(userID uint, authenticatedUser *models.User) error {
	// 权限检查：确保用户只能解绑自己的账号
	if userID != authenticatedUser.ID {
		slog.Warn("Permission denied for Google account unbinding", "userId", userID, "requesterId", authenticatedUser.ID)
		return errors.ErrPermissionDenied
	}

	// 检查是否已绑定Google账号
	_, err := s.userRepo.GetUserProvider(userID, "google")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrProviderNotBound
		}
		slog.Error("Failed to check user's Google provider", "userId", userID, "error", err)
		return fmt.Errorf("failed to check user provider: %w", err)
	}

	// 删除绑定记录
	if err := s.userRepo.DeleteUserProvider(userID, "google"); err != nil {
		slog.Error("Failed to unbind Google account", "userId", userID, "error", err)
		return fmt.Errorf("failed to unbind Google account: %w", err)
	}

	slog.Info("Google account unbound successfully", "userId", userID)
	return nil
}
