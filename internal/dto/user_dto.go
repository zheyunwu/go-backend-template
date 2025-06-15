package dto

import (
	"time"

	"github.com/go-backend-template/internal/models"
)

/* Response DTOs */

// UserProfileDTO represents user information.
type UserProfileDTO struct {
	ID              uint              `json:"id"`
	Name            string            `json:"name"`
	AvatarURL       *string           `json:"avatar_url"`
	Gender          models.GenderType `json:"gender"`
	Email           *string           `json:"email"`
	IsEmailVerified bool              `json:"is_email_verified"`
	Phone           *string           `json:"phone"`
	BirthDate       *time.Time        `json:"birth_date"`
	Locale          string            `json:"locale"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToUserProfileDTO converts a User model to a UserProfileDTO.
func ToUserProfileDTO(user *models.User) *UserProfileDTO {
	if user == nil {
		return nil
	}

	return &UserProfileDTO{
		ID:              user.ID,
		Name:            user.Name,
		AvatarURL:       user.AvatarURL,
		Gender:          user.Gender,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		Phone:           user.Phone,
		BirthDate:       user.BirthDate,
		Locale:          user.Locale,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

/* Request DTOs */

// UpdateProfileRequest is the request for updating user profile information.
type UpdateProfileRequest struct {
	Name      string            `json:"name" validate:"omitempty,min=1,max=50"`
	AvatarURL *string           `json:"avatar_url" validate:"omitempty,url"`
	Gender    models.GenderType `json:"gender" validate:"omitempty,oneof=male female other"`
	Email     *string           `json:"email" validate:"omitempty,email"`
	Phone     *string           `json:"phone" validate:"omitempty,e164"` // E.164 phone number format
	BirthDate *string           `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Locale    string            `json:"locale" validate:"omitempty,len=2"` // ISO 639-1 language code
}

// ToUpdatesMap converts the update request to a map of fields for updating.
func (r *UpdateProfileRequest) ToUpdatesMap() map[string]interface{} {
	updates := map[string]interface{}{}
	if r.Name != "" {
		updates["Name"] = r.Name
	}
	if r.AvatarURL != nil && *r.AvatarURL != "" {
		updates["AvatarURL"] = r.AvatarURL
	}
	if r.Gender != "" {
		updates["Gender"] = r.Gender
	}
	if r.Email != nil && *r.Email != "" {
		updates["Email"] = r.Email
	}
	if r.Phone != nil && *r.Phone != "" {
		updates["Phone"] = r.Phone
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *r.BirthDate) // Dereference r.BirthDate
		if err == nil {                                          // Only add if parsing is successful
			// Ensure the time part is zeroed out, keeping only the date part
			birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
			updates["BirthDate"] = birthDate
		}
	}
	if r.Locale != "" {
		updates["Locale"] = r.Locale
	}
	return updates
}

// UpdatePasswordRequest defines the DTO for updating a user's password.
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"` // Current password
	NewPassword     string `json:"new_password" validate:"required,min=8"`     // New password
}

// RegisterFromWechatMiniProgramRequest is the request for WeChat Mini Program registration.
type RegisterFromWechatMiniProgramRequest struct {
	UpdateProfileRequest
}

// ToModel converts a WeChat Mini Program registration request to a User model.
func (r *RegisterFromWechatMiniProgramRequest) ToModel() *models.User {
	user := models.User{
		Name:      r.Name,
		AvatarURL: r.AvatarURL,
		Gender:    r.Gender,
	}

	if r.Email != nil && *r.Email != "" {
		user.Email = r.Email
	}
	if r.Phone != nil && *r.Phone != "" {
		user.Phone = r.Phone
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *r.BirthDate)
		// Set birth date only if parsing is successful
		if err == nil {
			// Ensure the time part is zeroed out, keeping only the date part
			birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
			user.BirthDate = &birthDate
		}
	}
	if r.Locale != "" {
		user.Locale = r.Locale
	}

	return &user
}

// RegisterWithPasswordRequest is the request for registration using a password.
type RegisterWithPasswordRequest struct {
	UpdateProfileRequest
	Password string `json:"password" validate:"required,min=8"`
}

// ToModel converts a password registration request to a User model, including the hashed password.
func (r *RegisterWithPasswordRequest) ToModel(hashedPassword string) *models.User {
	user := models.User{
		Name:      r.Name,
		AvatarURL: r.AvatarURL,
		Gender:    r.Gender,
	}

	if r.Email != nil && *r.Email != "" {
		user.Email = r.Email
	}
	if r.Phone != nil && *r.Phone != "" {
		user.Phone = r.Phone
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *r.BirthDate)
		// Set birth date only if parsing is successful
		if err == nil {
			// Ensure the time part is zeroed out, keeping only the date part
			birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
			user.BirthDate = &birthDate
		}
	}
	if r.Locale != "" {
		user.Locale = r.Locale
	}

	if hashedPassword != "" {
		user.Password = &hashedPassword // Use the hashed password
	}

	return &user
}

// LoginWithPasswordRequest is the request for logging in with a password.
type LoginWithPasswordRequest struct {
	// Username or email
	EmailOrPhone string `json:"email_or_phone" validate:"required"` // Further validation can be handled in logic
	Password     string `json:"password" validate:"required"`
}

// RefreshTokenRequest is the request for refreshing an access token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"` // Refresh token
}

// DTOs related to email verification

// SendEmailVerificationRequest DTO for requesting email verification.
type SendEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"` // Email address
}

// VerifyEmailRequest DTO for verifying an email with a code.
type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"` // Email address
	Code  string `json:"code" validate:"required,len=6"`  // 6-digit verification code
}

// PasswordResetRequest DTO for requesting a password reset.
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"` // Email address
}

// PasswordResetConfirmRequest DTO for confirming a password reset.
type PasswordResetConfirmRequest struct {
	Email       string `json:"email" validate:"required,email"`        // Email address
	ResetToken  string `json:"reset_token" validate:"required,len=8"`  // 8-digit reset token
	NewPassword string `json:"new_password" validate:"required,min=8"` // New password
}

// DTOs related to OAuth2

// WechatOAuthRequest is the unified request for WeChat OAuth2 (login or registration).
type WechatOAuthRequest struct {
	Code       string `json:"code" validate:"required"`                      // WeChat OAuth authorization code
	ClientType string `json:"client_type" validate:"required,oneof=web app"` // Client type: web or app
}

// GoogleOAuthRequest is the unified request for Google OAuth2 (login or registration).
type GoogleOAuthRequest struct {
	Code         string `json:"code" validate:"required"`                      // OAuth authorization code
	CodeVerifier string `json:"code_verifier" validate:"required"`             // PKCE code verifier
	RedirectURI  string `json:"redirect_uri" validate:"required,url"`          // Redirect URI, must match configuration
	ClientType   string `json:"client_type" validate:"required,oneof=ios web"` // Client type: ios or web
}

// DTOs related to account binding

// BindWechatAccountRequest is the request for binding a WeChat account.
type BindWechatAccountRequest struct {
	WechatOAuthRequest // Directly use WechatOAuthRequest for binding request
}

// BindGoogleAccountRequest is the request for binding a Google account.
type BindGoogleAccountRequest struct {
	GoogleOAuthRequest // Directly use GoogleOAuthRequest for binding request
}
