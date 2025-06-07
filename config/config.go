package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`         // 数据库驱动类型，如 mysql, postgres
	Host         string `mapstructure:"host"`           // 数据库主机地址
	Port         int    `mapstructure:"port"`           // 数据库端口
	User         string `mapstructure:"user"`           // 数据库用户名
	Password     string `mapstructure:"password"`       // 数据库密码
	Name         string `mapstructure:"name"`           // 数据库名称
	Charset      string `mapstructure:"charset"`        // 数据库字符集
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 最大空闲连接数
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 最大打开连接数
}

// Config 结构体，映射到 YAML 配置
type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`

	Database DatabaseConfig `mapstructure:"database"`

	JWT struct {
		Secret      string `mapstructure:"secret"`
		ExpireHours int    `mapstructure:"expire_hours"`
	} `mapstructure:"jwt"`

	// Google OAuth2 配置
	Google struct {
		// iOS客户端配置
		IOS struct {
			ClientID     string   `mapstructure:"client_id"`
			ClientSecret string   `mapstructure:"client_secret"`
			RedirectURLs []string `mapstructure:"redirect_urls"`
		} `mapstructure:"ios"`
		// Web客户端配置
		Web struct {
			ClientID     string   `mapstructure:"client_id"`
			ClientSecret string   `mapstructure:"client_secret"`
			RedirectURLs []string `mapstructure:"redirect_urls"`
		} `mapstructure:"web"`
	} `mapstructure:"google"`

	// 微信OAuth2配置
	Wechat struct {
		Web struct {
			AppID  string `mapstructure:"appid"`
			Secret string `mapstructure:"secret"`
		} `mapstructure:"web"`
		App struct {
			AppID  string `mapstructure:"appid"`
			Secret string `mapstructure:"secret"`
		} `mapstructure:"app"`
	} `mapstructure:"wechat"`

	// 邮件服务配置
	Email struct {
		Provider       string `mapstructure:"provider"`         // 邮件服务提供商: sendgrid, smtp
		SendGridAPIKey string `mapstructure:"sendgrid_api_key"` // SendGrid API Key
		FromEmail      string `mapstructure:"from_email"`       // 发送邮箱地址
		FromName       string `mapstructure:"from_name"`        // 发送者名称
		// SMTP配置 (当provider为smtp时使用)
		SMTP struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			TLS      bool   `mapstructure:"tls"`
		} `mapstructure:"smtp"`
	} `mapstructure:"email"`

	// Redis配置
	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	// AI相关配置
	AI struct {
		OpenAIAPIKey   string `mapstructure:"openai_api_key"`
		MoonshotAPIURL string `mapstructure:"moonshot_api_url"`
		MoonshotAPIKey string `mapstructure:"moonshot_api_key"`
		DeepSeekAPIURL string `mapstructure:"deepseek_api_url"`
		DeepSeekAPIKey string `mapstructure:"deepseek_api_key"`
	} `mapstructure:"ai"`

	// 微信云托管相关配置
	WeChatCloudRun struct {
		Storage struct {
			COSBucket string `mapstructure:"cos_bucket"` // 云存储桶名称
			COSRegion string `mapstructure:"cos_region"` // 云存储区域，如 ap-shanghai
		} `mapstructure:"storage"`
	} `mapstructure:"wechat_cloudrun"`
}

func LoadConfig(env string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config." + env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AutomaticEnv()                                   // 支持环境变量覆盖
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 将点号替换为下划线，便于环境变量使用
	v.BindEnv("wechat_cloudrun.storage.cos_bucket", "COS_BUCKET")
	v.BindEnv("wechat_cloudrun.storage.cos_region", "COS_REGION")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &cfg, nil
}
