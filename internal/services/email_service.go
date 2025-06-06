package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"time"

	"github.com/go-backend-template/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService 邮件服务接口
type EmailService interface {
	// 发送邮箱验证邮件
	SendEmailVerification(to, name, verificationCode, locale string) error
	// 发送密码重置邮件
	SendPasswordReset(to, name, resetToken, locale string) error
}

// emailService 邮件服务实现
type emailService struct {
	config *config.Config
}

// EmailTemplate 邮件模板结构
type EmailTemplate struct {
	Subject     string
	HTMLContent string
	TextContent string
}

// 邮件验证模板
var emailVerificationTemplates = map[string]EmailTemplate{
	"zh": {
		Subject: "邮箱验证",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>邮箱验证</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .code { background: #e7f3ff; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ .AppName }}</h1>
        </div>
        <div class="content">
            <h2>邮箱验证</h2>
            <p>尊敬的 {{ .Name }}，</p>
            <p>感谢您注册我们的服务！请使用下面的验证码完成邮箱验证：</p>
            <div class="code">{{ .Code }}</div>
            <p>验证码有效期为 <strong>10 分钟</strong>，请尽快使用。</p>
            <p>如果您没有注册我们的服务，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>这是一封自动发送的邮件，请勿回复。</p>
            <p>&copy; {{ .Year }} {{ .AppName }}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `
邮箱验证 - {{ .AppName }}

尊敬的 {{ .Name }}，

感谢您注册我们的服务！请使用下面的验证码完成邮箱验证：

验证码：{{ .Code }}

验证码有效期为 10 分钟，请尽快使用。

如果您没有注册我们的服务，请忽略此邮件。

这是一封自动发送的邮件，请勿回复。
© {{ .Year }} {{ .AppName }}. All rights reserved.`,
	},
	"en": {
		Subject: "Email Verification",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .code { background: #e7f3ff; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ .AppName }}</h1>
        </div>
        <div class="content">
            <h2>Email Verification</h2>
            <p>Dear {{ .Name }},</p>
            <p>Thank you for registering with our service! Please use the verification code below to complete your email verification:</p>
            <div class="code">{{ .Code }}</div>
            <p>The verification code is valid for <strong>10 minutes</strong>. Please use it promptly.</p>
            <p>If you did not register for our service, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated email, please do not reply.</p>
            <p>&copy; {{ .Year }} {{ .AppName }}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `
Email Verification - {{ .AppName }}

Dear {{ .Name }},

Thank you for registering with our service! Please use the verification code below to complete your email verification:

Verification Code: {{ .Code }}

The verification code is valid for 10 minutes. Please use it promptly.

If you did not register for our service, please ignore this email.

This is an automated email, please do not reply.
© {{ .Year }} {{ .AppName }}. All rights reserved.`,
	},
	"de": {
		Subject: "E-Mail-Verifizierung",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>E-Mail-Verifizierung</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .code { background: #e7f3ff; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ .AppName }}</h1>
        </div>
        <div class="content">
            <h2>E-Mail-Verifizierung</h2>
            <p>Liebe/r {{ .Name }},</p>
            <p>Vielen Dank für Ihre Registrierung bei unserem Service! Bitte verwenden Sie den folgenden Verifizierungscode, um Ihre E-Mail-Adresse zu bestätigen:</p>
            <div class="code">{{ .Code }}</div>
            <p>Der Verifizierungscode ist <strong>10 Minuten</strong> gültig. Bitte verwenden Sie ihn umgehend.</p>
            <p>Falls Sie sich nicht für unseren Service registriert haben, ignorieren Sie bitte diese E-Mail.</p>
        </div>
        <div class="footer">
            <p>Dies ist eine automatisch generierte E-Mail, bitte antworten Sie nicht darauf.</p>
            <p>&copy; {{ .Year }} {{ .AppName }}. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `
E-Mail-Verifizierung - {{ .AppName }}

Liebe/r {{ .Name }},

Vielen Dank für Ihre Registrierung bei unserem Service! Bitte verwenden Sie den folgenden Verifizierungscode, um Ihre E-Mail-Adresse zu bestätigen：

Verifizierungscode: {{ .Code }}

Der Verifizierungscode ist 10 Minuten gültig. Bitte verwenden Sie ihn umgehend.

Falls Sie sich nicht für unseren Service registriert haben, ignorieren Sie bitte diese E-Mail.

Dies ist eine automatisch generierte E-Mail, bitte antworten Sie nicht darauf.
© {{ .Year }} {{ .AppName }}. Alle Rechte vorbehalten.`,
	},
}

// 密码重置模板
var passwordResetTemplates = map[string]EmailTemplate{
	"zh": {
		Subject: "密码重置",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>密码重置</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .code { background: #fff3cd; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ .AppName }}</h1>
        </div>
        <div class="content">
            <h2>密码重置</h2>
            <p>尊敬的 {{ .Name }}，</p>
            <p>您请求重置密码。请使用下面的重置码：</p>
            <div class="code">{{ .Code }}</div>
            <p>重置码有效期为 <strong>30 分钟</strong>，请尽快使用。</p>
            <p>如果您没有请求重置密码，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>这是一封自动发送的邮件，请勿回复。</p>
            <p>&copy; {{ .Year }} {{ .AppName }}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `
密码重置 - {{ .AppName }}

尊敬的 {{ .Name }}，

您请求重置密码。请使用下面的重置码：

重置码：{{ .Code }}

重置码有效期为 30 分钟，请尽快使用。

如果您没有请求重置密码，请忽略此邮件。

这是一封自动发送的邮件，请勿回复。
© {{ .Year }} {{ .AppName }}. All rights reserved.`,
	},
	"en": {
		Subject: "Password Reset",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .code { background: #fff3cd; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ .AppName }}</h1>
        </div>
        <div class="content">
            <h2>Password Reset</h2>
            <p>Dear {{ .Name }},</p>
            <p>You have requested to reset your password. Please use the reset code below:</p>
            <div class="code">{{ .Code }}</div>
            <p>The reset code is valid for <strong>30 minutes</strong>. Please use it promptly.</p>
            <p>If you did not request a password reset, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated email, please do not reply.</p>
            <p>&copy; {{ .Year }} {{ .AppName }}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `
Password Reset - {{ .AppName }}

Dear {{ .Name }},

You have requested to reset your password. Please use the reset code below:

Reset Code: {{ .Code }}

The reset code is valid for 30 minutes. Please use it promptly.

If you did not request a password reset, please ignore this email.

This is an automated email, please do not reply.
© {{ .Year }} {{ .AppName }}. All rights reserved.`,
	},
	"de": {
		Subject: "Passwort zurücksetzen",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Passwort zurücksetzen</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .code { background: #fff3cd; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ .AppName }}</h1>
        </div>
        <div class="content">
            <h2>Passwort zurücksetzen</h2>
            <p>Liebe/r {{ .Name }},</p>
            <p>Sie haben eine Passwort-Zurücksetzung angefordert. Bitte verwenden Sie den folgenden Zurücksetzungscode:</p>
            <div class="code">{{ .Code }}</div>
            <p>Der Zurücksetzungscode ist <strong>30 Minuten</strong> gültig. Bitte verwenden Sie ihn umgehend.</p>
            <p>Falls Sie keine Passwort-Zurücksetzung angefordert haben, ignorieren Sie bitte diese E-Mail.</p>
        </div>
        <div class="footer">
            <p>Dies ist eine automatisch generierte E-Mail, bitte antworten Sie nicht darauf.</p>
            <p>&copy; {{ .Year }} {{ .AppName }}. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `
Passwort zurücksetzen - {{ .AppName }}

Liebe/r {{ .Name }},

Sie haben eine Passwort-Zurücksetzung angefordert. Bitte verwenden Sie den folgenden Zurücksetzungscode:

Zurücksetzungscode: {{ .Code }}

Der Zurücksetzungscode ist 30 Minuten gültig. Bitte verwenden Sie ihn umgehend.

Falls Sie keine Passwort-Zurücksetzung angefordert haben, ignorieren Sie bitte diese E-Mail.

Dies ist eine automatisch generierte E-Mail, bitte antworten Sie nicht darauf.
© {{ .Year }} {{ .AppName }}. Alle Rechte vorbehalten.`,
	},
}

// NewEmailService 创建邮件服务实例
func NewEmailService(config *config.Config) EmailService {
	return &emailService{
		config: config,
	}
}

// getSupportedLanguage 从locale获取支持的语言，遵循IETF BCP 47标准
func (s *emailService) getSupportedLanguage(locale string) string {
	// 支持的语言映射表
	supportedLanguages := map[string]string{
		// 中文相关的locale
		"zh":    "zh",
		"zh-CN": "zh",
		"zh-TW": "zh",
		"zh-HK": "zh",
		"zh-SG": "zh",
		// 英文相关的locale
		"en":    "en",
		"en-US": "en",
		"en-GB": "en",
		"en-AU": "en",
		"en-CA": "en",
		// 德文相关的locale
		"de":    "de",
		"de-DE": "de",
		"de-AT": "de",
		"de-CH": "de",
	}

	// 直接匹配完整locale
	if lang, exists := supportedLanguages[locale]; exists {
		return lang
	}

	// 如果完整locale不匹配，尝试匹配语言部分（取'-'前面的部分）
	if len(locale) >= 2 {
		langCode := locale[:2]
		if lang, exists := supportedLanguages[langCode]; exists {
			return lang
		}
	}

	// 默认返回英文
	return "en"
}

// SendEmailVerification 发送邮箱验证邮件
func (s *emailService) SendEmailVerification(to, name, verificationCode, locale string) error {
	// 验证并获取支持的语言
	lang := s.getSupportedLanguage(locale)

	// 获取对应语言的模板
	template := emailVerificationTemplates[lang]

	subject := template.Subject + " - " + s.config.Email.FromName

	data := struct {
		Name    string
		Code    string
		AppName string
		Year    int
	}{
		Name:    name,
		Code:    verificationCode,
		AppName: s.config.Email.FromName,
		Year:    time.Now().Year(),
	}

	htmlContent, err := s.renderTemplate(template.HTMLContent, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML email template: %w", err)
	}

	textContent, err := s.renderTemplate(template.TextContent, data)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	return s.sendEmail(to, subject, textContent, htmlContent)
}

// SendPasswordReset 发送密码重置邮件
func (s *emailService) SendPasswordReset(to, name, resetToken, locale string) error {
	// 验证并获取支持的语言
	lang := s.getSupportedLanguage(locale)

	// 获取对应语言的模板
	template := passwordResetTemplates[lang]

	subject := template.Subject + " - " + s.config.Email.FromName

	data := struct {
		Name    string
		Code    string
		AppName string
		Year    int
	}{
		Name:    name,
		Code:    resetToken,
		AppName: s.config.Email.FromName,
		Year:    time.Now().Year(),
	}

	htmlContent, err := s.renderTemplate(template.HTMLContent, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML email template: %w", err)
	}

	textContent, err := s.renderTemplate(template.TextContent, data)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	return s.sendEmail(to, subject, textContent, htmlContent)
}

// sendEmail 根据配置选择邮件发送方式
func (s *emailService) sendEmail(to, subject, textContent, htmlContent string) error {
	switch s.config.Email.Provider {
	case "sendgrid":
		return s.sendWithSendGrid(to, subject, textContent, htmlContent)
	case "smtp":
		return s.sendWithSMTP(to, subject, textContent, htmlContent)
	default:
		return fmt.Errorf("unsupported email provider: %s", s.config.Email.Provider)
	}
}

// sendWithSendGrid 使用 SendGrid 发送邮件
func (s *emailService) sendWithSendGrid(to, subject, textContent, htmlContent string) error {
	from := mail.NewEmail(s.config.Email.FromName, s.config.Email.FromEmail)
	toEmail := mail.NewEmail("", to)

	message := mail.NewSingleEmail(from, subject, toEmail, textContent, htmlContent)

	client := sendgrid.NewSendClient(s.config.Email.SendGridAPIKey)
	response, err := client.Send(message)

	if err != nil {
		slog.Error("Failed to send email via SendGrid", "error", err, "to", to)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		slog.Error("SendGrid returned error", "status", response.StatusCode, "body", response.Body, "to", to)
		return fmt.Errorf("sendgrid error: status %d", response.StatusCode)
	}

	slog.Info("Email sent successfully via SendGrid", "to", to, "status", response.StatusCode)
	return nil
}

// sendWithSMTP 使用 SMTP 发送邮件
func (s *emailService) sendWithSMTP(to, subject, textContent, htmlContent string) error {
	auth := smtp.PlainAuth("", s.config.Email.SMTP.Username, s.config.Email.SMTP.Password, s.config.Email.SMTP.Host)

	// 构建邮件内容
	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.Email.FromName, s.config.Email.FromEmail))
	body.WriteString(fmt.Sprintf("To: %s\r\n", to))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString("MIME-Version: 1.0\r\n")

	if htmlContent != "" {
		body.WriteString("Content-Type: multipart/alternative; boundary=\"boundary\"\r\n\r\n")
		body.WriteString("--boundary\r\n")
		body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		body.WriteString(textContent + "\r\n")
		body.WriteString("--boundary\r\n")
		body.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		body.WriteString(htmlContent + "\r\n")
		body.WriteString("--boundary--\r\n")
	} else {
		body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		body.WriteString(textContent)
	}

	addr := fmt.Sprintf("%s:%d", s.config.Email.SMTP.Host, s.config.Email.SMTP.Port)
	err := smtp.SendMail(addr, auth, s.config.Email.FromEmail, []string{to}, body.Bytes())

	if err != nil {
		slog.Error("Failed to send email via SMTP", "error", err, "to", to)
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("Email sent successfully via SMTP", "to", to)
	return nil
}

// renderTemplate 渲染邮件模板
func (s *emailService) renderTemplate(templateContent string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
