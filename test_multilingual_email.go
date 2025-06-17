package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/services"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("dev")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建邮件服务
	emailService := services.NewEmailService(cfg)

	// 测试用户信息
	testEmail := "wuzheyun@gmail.com"
	testName := "张三"
	verificationCode := "123456"
	resetToken := "ABCD1234"

	fmt.Println("=== 测试多语言邮件发送 ===")
	// 测试中文邮箱验证
	fmt.Println("\n1. 测试中文邮箱验证邮件")
	err = emailService.SendEmailVerification(context.Background(), testEmail, testName, verificationCode, "zh-CN")
	if err != nil {
		fmt.Printf("发送中文验证邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 中文验证邮件发送成功")
	}
	// 测试英文邮箱验证
	fmt.Println("\n2. 测试英文邮箱验证邮件")
	err = emailService.SendEmailVerification(context.Background(), testEmail, "John Doe", verificationCode, "en-US")
	if err != nil {
		fmt.Printf("发送英文验证邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 英文验证邮件发送成功")
	}
	// 测试德文邮箱验证
	fmt.Println("\n3. 测试德文邮箱验证邮件")
	err = emailService.SendEmailVerification(context.Background(), testEmail, "Hans Müller", verificationCode, "de-DE")
	if err != nil {
		fmt.Printf("发送德文验证邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 德文验证邮件发送成功")
	}
	// 测试不支持的语言（应该回退到英文）
	fmt.Println("\n4. 测试不支持的语言（回退到英文）")
	err = emailService.SendEmailVerification(context.Background(), testEmail, testName, verificationCode, "fr-FR")
	if err != nil {
		fmt.Printf("发送回退语言邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 回退语言邮件发送成功（应为英文）")
	}
	// 测试中文密码重置
	fmt.Println("\n5. 测试中文密码重置邮件")
	err = emailService.SendPasswordReset(context.Background(), testEmail, testName, resetToken, "zh-CN")
	if err != nil {
		fmt.Printf("发送中文重置邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 中文重置邮件发送成功")
	}
	// 测试英文密码重置
	fmt.Println("\n6. 测试英文密码重置邮件")
	err = emailService.SendPasswordReset(context.Background(), testEmail, "John Doe", resetToken, "en-US")
	if err != nil {
		fmt.Printf("发送英文重置邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 英文重置邮件发送成功")
	}
	// 测试德文密码重置
	fmt.Println("\n7. 测试德文密码重置邮件")
	err = emailService.SendPasswordReset(context.Background(), testEmail, "Hans Müller", resetToken, "de-DE")
	if err != nil {
		fmt.Printf("发送德文重置邮件失败: %v\n", err)
	} else {
		fmt.Println("✓ 德文重置邮件发送成功")
	}

	fmt.Println("\n=== 多语言邮件测试完成 ===")
	fmt.Println("\n注意：")
	fmt.Println("- 请确保在 config.dev.yaml 中正确配置了邮件服务")
	fmt.Println("- 如果使用 SMTP，请确保 SMTP 服务器配置正确")
	fmt.Println("- 如果使用 SendGrid，请确保 API Key 配置正确")
	fmt.Println("- 检查收件箱中是否收到了不同语言的邮件")
}
