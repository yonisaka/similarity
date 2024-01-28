package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/yonisaka/similarity/internal/usecases"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var code int

	defer func() {
		os.Exit(code)
	}()

	_ = os.Setenv("OPENAI_API_KEY", "test")

	code = m.Run()
}

func TestSearch(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OpenAI API key is not set")
		return
	}

	client := openai.NewClient(apiKey)

	searchUsecase := usecases.NewSearchUsecase(client)
	ctx := context.Background()
	answer, err := searchUsecase.Search(ctx, "stok nomor BA00001323K14 memiliki plat nomor apa?")
	if err != nil {
		log.Println(err)
		assert.Error(t, err)
		return
	}

	assert.Contains(t, answer, "B1207KDZ")
	log.Println(answer)
}
