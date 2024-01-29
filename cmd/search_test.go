package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yonisaka/similarity/internal/di"
	"log"
	"testing"
)

func TestSearch(t *testing.T) {
	searchUsecase := di.GetSearchUsecase()

	qna1 := make(map[string]string)
	//qna1["stok nomor BA00002123J16 dimiliki penjual apa?"] = "Yuliana"
	qna1["stok nomor BA00001323K14 memiliki plat nomor apa?"] = "B1207KDZ"
	//qna1["mobil dengan plat nomor F1088DA memiliki warna apa?"] = "Hitam Metalic"

	qna2 := make(map[string]string)
	//qna2["mobil dengan plat nomor B1690PRD memiliki harga awal berapa?"] = "62000000"
	qna2["mobil dengan plat nomor B1207KDZ memiliki segment apa?"] = "Sedan"

	qna3 := make(map[string]string)
	//qna3["stok nomor BA00001023J09 memiliki alamat penjual dimana?"] = "Grand Taruma"
	qna3["mobil dengan plat nomor B9910KYX memiliki pabrikan apa?"] = "Mercedes"

	for question, expectedAnswerContains := range qna1 {
		ctx := context.Background()
		result, err := searchUsecase.Search(ctx, question)
		if err != nil {
			log.Println(err)
		}

		log.Println(
			fmt.Sprintf(
				"\nquestion: %s \n answer: %s \n expected: %s \n", question, result, expectedAnswerContains,
			),
		)

		assert.Contains(t, result, expectedAnswerContains)
	}
}
