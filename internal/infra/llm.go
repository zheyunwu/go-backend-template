package infra

import (
	"github.com/go-backend-template/config"
	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"
)

func InitOpenAIClient(config *config.Config) *openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(config.AI.MoonshotAPIKey),
		option.WithBaseURL(config.AI.MoonshotAPIURL),
	)

	return &client
}

func InitMoonshotClient(config *config.Config) *openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(config.AI.MoonshotAPIKey),
		option.WithBaseURL(config.AI.MoonshotAPIURL),
	)

	return &client
}

func InitDeepSeekClient(config *config.Config) *openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(config.AI.DeepSeekAPIKey),
		option.WithBaseURL(config.AI.DeepSeekAPIURL),
	)

	return &client
}
