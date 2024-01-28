package di

import (
	"github.com/sashabaranov/go-openai"
	"os"
)

// GetOpenAIClient returns OpenAI client instance.
func GetOpenAIClient() openai.Client {
	return *openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}
