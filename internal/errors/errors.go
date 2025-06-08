package errors

import (
	"errors"
	"net/http"
)

// 定义应用错误
var (
	// 通用错误
	ErrInternalServer     = NewAppError("internal_server_error", "Internal server error", http.StatusInternalServerError)
	ErrBadRequest         = NewAppError("bad_request", "Bad request", http.StatusBadRequest)
	ErrNotFound           = NewAppError("not_found", "Resource not found", http.StatusNotFound)
	ErrUnauthorized       = NewAppError("unauthorized", "Unauthorized", http.StatusUnauthorized)
	ErrPermissionDenied   = NewAppError("permission_denied", "Permission denied", http.StatusForbidden)
	ErrTooManyRequests    = NewAppError("too_many_requests", "Too many requests", http.StatusTooManyRequests)
	ErrInvalidCredentials = NewAppError("invalid_credentials", "Invalid credentials", http.StatusUnauthorized)
	ErrNoValidUpdates     = NewAppError("no_valid_updates", "No valid fields to update", http.StatusBadRequest)
	ErrInvalidDateFormat  = NewAppError("invalid_date_format", "Invalid date format", http.StatusBadRequest)

	// 用户相关错误
	ErrUserNotFound            = NewAppError("user_not_found", "User not found", http.StatusNotFound)
	ErrUserAlreadyExists       = NewAppError("user_already_exists", "User already exists", http.StatusConflict)
	ErrPhoneAlreadyExists      = NewAppError("phone_already_exists", "Phone number already exists", http.StatusConflict)
	ErrEmailAlreadyExists      = NewAppError("email_already_exists", "Email already exists", http.StatusConflict)
	ErrEmailOrPhoneNotProvided = NewAppError("email_or_phone_not_provided", "Email or phone number not provided", http.StatusBadRequest)
	ErrInvalidEmail            = NewAppError("invalid_email", "Invalid email format", http.StatusBadRequest)
	ErrInvalidPassword         = NewAppError("invalid_password", "Invalid password", http.StatusBadRequest)
	ErrPasswordTooShort        = NewAppError("password_too_short", "Password must be at least 8 characters", http.StatusBadRequest)
	ErrPasswordTooWeak         = NewAppError("password_too_weak", "Password is too weak", http.StatusBadRequest)
	ErrUserBanned              = NewAppError("user_banned", "User is banned", http.StatusForbidden)
	ErrInvalidToken            = NewAppError("invalid_token", "Invalid or expired token", http.StatusUnauthorized)
	ErrInvalidVerificationCode = NewAppError("invalid_verification_code", "Invalid verification code", http.StatusBadRequest)
	ErrVerificationCodeExpired = NewAppError("verification_code_expired", "Verification code expired", http.StatusBadRequest)

	// 邮箱验证相关错误
	ErrEmailNotVerified            = NewAppError("email_not_verified", "Email address is not verified", http.StatusUnauthorized)
	ErrEmailAlreadyVerified        = NewAppError("email_already_verified", "Email address is already verified", http.StatusBadRequest)
	ErrTooManyVerificationRequests = NewAppError("too_many_verification_requests", "Too many verification requests. Please wait before requesting again", http.StatusTooManyRequests)

	// 账号绑定相关错误
	ErrProviderAlreadyBound = NewAppError("provider_already_bound", "Account is already bound to this or another user", http.StatusConflict)
	ErrProviderNotBound     = NewAppError("provider_not_bound", "Account is not bound", http.StatusNotFound)

	// 商品相关错误
	ErrProductNotFound   = NewAppError("product_not_found", "Product not found", http.StatusNotFound)
	ErrProductNameEmpty  = NewAppError("product_name_empty", "Product name cannot be empty", http.StatusBadRequest)
	ErrInvalidBarcode    = NewAppError("invalid_barcode", "Invalid barcode", http.StatusBadRequest)
	ErrBarcodeExists     = NewAppError("barcode_exists", "Product with this barcode already exists", http.StatusConflict)
	ErrCategoryNotFound  = NewAppError("category_not_found", "Category not found", http.StatusNotFound)
	ErrProductImageEmpty = NewAppError("product_image_empty", "Product image cannot be empty", http.StatusBadRequest)

	// 审核相关错误
	ErrModeratorNotFound = NewAppError("moderator_not_found", "Moderator not found", http.StatusNotFound)

	// 评论相关错误
	ErrReviewNotFound  = NewAppError("review_not_found", "Review not found", http.StatusNotFound)
	ErrInvalidRating   = NewAppError("invalid_rating", "Rating must be between 1.0 and 5.0", http.StatusBadRequest)
	ErrDuplicateReview = NewAppError("duplicate_review", "User already reviewed this product", http.StatusConflict)

	// AI Quota相关错误
	ErrDailyTokenQuotaExceeded = NewAppError("daily_token_quota_exceeded", "Daily token quota exceeded", http.StatusTooManyRequests)
	ErrMonthlyRequestsExceeded = NewAppError("monthly_requests_exceeded", "Monthly requests quota exceeded", http.StatusTooManyRequests)

	// AI识别相关错误
	ErrInvalidImageFormat  = NewAppError("invalid_image_format", "Invalid image format", http.StatusBadRequest)
	ErrImageSizeExceeded   = NewAppError("image_size_exceeded", "Image size exceeds the limit", http.StatusBadRequest)
	ErrAIModelNotAvailable = NewAppError("ai_model_not_available", "AI model is not available", http.StatusServiceUnavailable)
	ErrNoRecognitionResult = NewAppError("no_recognition_result", "No recognition result", http.StatusNotFound)

	// Google OAuth2相关错误
	ErrInvalidOAuthCode         = NewAppError("invalid_oauth_code", "Invalid OAuth authorization code", http.StatusBadRequest)
	ErrOAuthTokenExchange       = NewAppError("oauth_token_exchange", "Failed to exchange OAuth code for token", http.StatusBadRequest)
	ErrOAuthUserInfoFetch       = NewAppError("oauth_user_info_fetch", "Failed to fetch user info from OAuth provider", http.StatusBadRequest)
	ErrInvalidRedirectURL       = NewAppError("invalid_redirect_url", "Invalid redirect URL", http.StatusBadRequest)
	ErrInvalidClientType        = NewAppError("invalid_client_type", "Invalid client type for OAuth", http.StatusBadRequest)
	ErrGoogleUserInfoIncomplete = NewAppError("google_user_info_incomplete", "Google user info is incomplete", http.StatusBadRequest)

	// 反馈相关错误
	ErrFeedbackNotFound         = NewAppError("feedback_not_found", "Feedback not found", http.StatusNotFound)
	ErrInvalidFeedback          = NewAppError("invalid_feedback", "Invalid feedback content", http.StatusBadRequest)
	ErrFeedbackCannotBeModified = NewAppError("feedback_cannot_be_modified", "Feedback cannot be modified in its current state", http.StatusBadRequest)
	ErrFeedbackCannotBeDeleted  = NewAppError("feedback_cannot_be_deleted", "Feedback cannot be deleted in its current state", http.StatusBadRequest)
	ErrInvalidStatusTransition  = NewAppError("invalid_status_transition", "Invalid feedback status transition", http.StatusBadRequest)

	// 内容安全检查相关错误
	ErrContentSecurityCheck = NewAppError("content_security_check", "包含不恰当内容", http.StatusBadRequest)
	ErrContentSecurityAPI   = NewAppError("content_security_api_error", "Content security check API error", http.StatusServiceUnavailable)
	ErrContentTooLong       = NewAppError("content_too_long", "内容过长", http.StatusBadRequest)
	ErrContentEmptyCheck    = NewAppError("content_empty", "Content cannot be empty", http.StatusBadRequest)
	ErrOpenIDRequired       = NewAppError("openid_required", "OpenID is required for content security check", http.StatusBadRequest)
)

// AppError 定义应用自定义错误
type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

// Error 实现error接口
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap 方法使用于 errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建新的应用错误
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     errors.New(message),
	}
}

// WithDetails 添加自定义错误细节
func (e *AppError) WithDetails(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Err:     err,
	}
}
