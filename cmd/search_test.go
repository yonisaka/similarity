package main

import (
	"context"
	"fmt"
	"github.com/yonisaka/similarity/internal/di"
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
	_ = os.Setenv("APP_ENV", "test")
	_ = os.Setenv("IS_REPLICA", "false")

	code = m.Run()
}

func TestSearch(t *testing.T) {
	searchUsecase := di.GetSearchUsecase()

	qna1 := make(map[string]string)
	qna1["stok nomor BA00002123J16 dimiliki penjual apa?"] = "Yuliana adec "
	qna1["stok nomor BA00001323K14 memiliki plat nomor apa?"] = "B1207KDZ"
	qna1["mobil dengan plat nomor F1088DA memiliki warna apa?"] = "Hitam Metalic"

	qna2 := make(map[string]string)
	qna2["mobil dengan plat nomor B1690PRD memiliki harga awal berapa?"] = "62000000"
	qna2["mobil dengan plat nomor B1207KDZ memiliki segment apa?"] = "Sedan"

	for question, expectedAnswerContains := range qna1 {
		ctx := context.Background()
		result, err := searchUsecase.Search(ctx, question)
		if err != nil {
			return
		}

		log.Println(
			fmt.Sprintf(
				"question: %s \n answer: %s \n expected: %s \n", question, result, expectedAnswerContains,
			),
		)
	}
}
