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

	AI struct {
		OpenAIAPIKey   string `mapstructure:"openai_api_key"`
		MoonshotAPIURL string `mapstructure:"moonshot_api_url"`
		MoonshotAPIKey string `mapstructure:"moonshot_api_key"`
		DeepSeekAPIURL string `mapstructure:"deepseek_api_url"`
		DeepSeekAPIKey string `mapstructure:"deepseek_api_key"`
	} `mapstructure:"ai"`

	// 微信云托管相关配置
	Cloud struct {
		COSBucket string `mapstructure:"cos_bucket" env:"COS_BUCKET"` // 云存储桶名称
		COSRegion string `mapstructure:"cos_region" env:"COS_REGION"` // 云存储区域，如 ap-shanghai
	} `mapstructure:"cloud"`
}

func LoadConfig(env string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config." + env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AutomaticEnv() // 支持环境变量覆盖
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.BindEnv("cloud.cos_bucket", "COS_BUCKET")
	v.BindEnv("cloud.cos_region", "COS_REGION")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &cfg, nil
}
