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

	qna := make(map[string]string, 4)
	qna["stok nomor BA00002123J16 dimiliki penjual apa?"] = "Yuliana adec "
	qna["stok nomor BA00001323K14 memiliki plat nomor apa?"] = "B1207KDZ"
	qna["mobil dengan plat nomor F1088DA memiliki warna apa?"] = "Hitam Metalic"
	qna["mobil dengan plat nomor B1690PRD memiliki harga awal berapa?"] = "62000000"

	for question, expectedAnswerContains := range qna {
		result, err := searchUsecase.Search(ctx, question)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println(
			fmt.Sprintf(
				"question: %s \n answer: %s \n expected: %s \n", question, result, expectedAnswerContains,
			),
		)

		assert.Contains(t, result, expectedAnswerContains)
	}
}
