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
	qna1["stok nomor BA00002123J16 dimiliki penjual apa?"] = "Yuliana" // ada:15 small:15
	//qna1["stok nomor BA00001323K14 memiliki plat nomor apa?"] = "B1207KDZ"        // ada:4 small:2
	//qna1["mobil dengan plat nomor F1088DA memiliki warna apa?"] = "Hitam Metalic" // ada:6 small:15

	qna2 := make(map[string]string)
	qna2["mobil dengan plat nomor T8324AP memiliki harga awal berapa?"] = "135000000"  // ada:1 small:1
	qna2["mobil dengan plat nomor B1207KDZ memiliki kapasitas mesin berapa?"] = "2496" // ada:3 small:5

	qna3 := make(map[string]string)
	qna3["stok nomor BA00001023J09 memiliki odometer berapa?"] = "108585" // ada:1 small:7
	//qna3["mobil dengan plat nomor D1167AGX memiliki nomor mesin apa?"] = "1NRF437889" // ada:2 small:1

	qna4 := make(map[string]string)
	qna4["stok nomor KA00000823G27 memiliki nomor rangka apa?"] = "MHMFM517BCK004179" // ada:4 small:11
	qna4["mobil dengan plat nomor B9721KYX memiliki pabrikan apa?"] = "Mercedes"      // ada:9 small:13

	qna5 := make(map[string]string)
	qna5["mobil dengan plat nomor D1167AGX memiliki segment apa?"] = "MPV"         // ada:1 small:1
	qna5["mobil dengan plat nomor BA9244LL memiliki pabrikan apa?"] = "Mitsubishi" // ada:6 small:9

	qna6 := make(map[string]string)
	qna6["stok nomor KA00000123C21 memiliki nomor mesin apa?"] = "MA92076"
	qna6["mobil dengan plat nomor D1040UW memiliki tipe bahan bakar apa?"] = "Bensin"

	qna7 := make(map[string]string)
	qna7["mobil dengan plat nomor B9195QZ memiliki grade apa?"] = "D"
	qna7["stok nomor BA00001423F23 memiliki spesifikasi apa?"] = "1497"

	qna8 := make(map[string]string)
	qna8["stok nomor BA00000123B13 berada di cabang apa?"] = "Bekasi"
	qna8["mobil dengan plat nomor B2396UFF memiliki model apa?"] = "MOBILIO"
	//qna8["cara membuat crud golang?"] = "golang"

	for question, expectedAnswerContains := range qna7 {
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
