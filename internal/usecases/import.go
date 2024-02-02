package usecases

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	pb "github.com/qdrant/go-client/qdrant"
	"github.com/yonisaka/similarity/internal/entities"
	"github.com/yonisaka/similarity/internal/entities/repository"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	maxTokens = 2000
)

var errGetEmbedding = errors.New("error getting embedding")

func (u *importUsecase) Import(ctx context.Context, filename string) error {
	combined, rawVectors, err := u.ReadCSV(filename)
	if err != nil {
		return err
	}

	offset, err := u.embeddingRepo.CountEmbeddingByScope(ctx, filename)
	if err != nil {
		return err
	}

	u.logger.Info(fmt.Sprintf("offset: %d", offset))

	tokens := 0
	for i, rawVector := range rawVectors {
		if i < offset {
			continue
		}

		if tokens > maxTokens {
			break
		}

		// Get the embedding for the combine
		embedding, nTokens, err := u.GetEmbedding(rawVector)
		if err != nil {
			if errors.Is(err, errGetEmbedding) {
				continue
			}
			return err
		}

		// Save the embedding to the database
		if err := u.embeddingRepo.CreateEmbedding(ctx, &repository.Embedding{
			Scope:     filename,
			Combined:  combined[i],
			Embedding: embedding,
			NTokens:   nTokens,
		}); err != nil {
			return err
		}

		tokens += nTokens

		u.logger.Info(fmt.Sprintf("usage tokens: %d", tokens))
	}

	return nil
}

func (u *importUsecase) ReadCSV(filename string) ([]string, []string, error) {
	// Open the CSV file
	file, err := os.Open(fmt.Sprintf("../data/%s", filename))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new CSV reader reading from the opened file
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	// Assuming the first line is headers
	headers, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	// Read the rest of the data
	var records [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				log.Println("Warning: Skipping a malformed line")
				continue
			} else if err == io.EOF {
				// End of file is reached
				break
			}

			break
		}

		records = append(records, record)
	}

	var combined []string
	var rawVectors []string
	for _, record := range records {
		combine := ""
		rawVector := ""
		for i, field := range record {
			if field == "" {
				continue
			}

			if i == 0 {
				combine = fmt.Sprintf("%s: %s", headers[i], field)
				rawVector = field
			} else {
				combine = fmt.Sprintf("%s; %s: %s", combine, headers[i], field)
				rawVector = fmt.Sprintf("%s %s", rawVector, field)
			}
		}

		combined = append(combined, combine)
		rawVectors = append(rawVectors, rawVector)
	}

	return combined, rawVectors, nil
}

func (u *importUsecase) GetEmbedding(query string) ([]float64, int, error) {
	// Construct the request body
	requestBody, err := json.Marshal(map[string]string{
		"input": query,
		"model": os.Getenv("OPENAI_EMBEDDING_MODEL"),
	})
	if err != nil {
		u.logger.Warn(fmt.Sprintf("Error occurred while marshaling. %s", err))
		return nil, 0, err
	}

	// Create an HTTP request
	req, err := http.NewRequest("POST", embeddingURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		u.logger.Warn(fmt.Sprintf("Error occurred while making HTTP request. %s", err))
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error occurred while reading response body. %s", err)
	}

	// Unmarshal the response into the EmbeddingResponse struct
	var embeddingResponse *entities.EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		return nil, 0, err
	}

	// Return the embedding
	if len(embeddingResponse.Data) > 0 {
		return embeddingResponse.Data[0].Embedding, embeddingResponse.Usage.PromptTokens, nil
	}

	var errorResponse *entities.ErrorResponse
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return nil, 0, err
	}

	if errorResponse != nil {
		return nil, 0, errors.New(errorResponse.Error.Message)
	}

	return nil, 0, errGetEmbedding
}

func (u *importUsecase) MigrateToQdrant(ctx context.Context) error {
	records, err := u.embeddingRepo.ListEmbeddingByScope(ctx, "sample_lelang.csv")
	if err != nil {
		return err
	}

	if err := u.qdrantClient.DeleteCollection(u.qdrantClient.GetCollectionName()); err != nil {
		return err
	}

	if err := u.qdrantClient.CreateCollection(
		u.qdrantClient.GetCollectionName(),
		u.qdrantClient.GetVectorSize(),
	); err != nil {
		return err
	}

	var points []*pb.PointStruct
	for _, record := range records {
		point := u.buildPoint(record.Combined, convertToFloat32(record.Embedding))
		points = append(points, point)
	}

	return u.qdrantClient.CreatePoints(points)
}

func (u *importUsecase) buildPoint(combined string, embedding []float32) *pb.PointStruct {
	point := &pb.PointStruct{}

	point.Id = &pb.PointId{
		PointIdOptions: &pb.PointId_Uuid{
			Uuid: md5str(combined),
		},
	}

	// vector
	point.Vectors = &pb.Vectors{
		VectorsOptions: &pb.Vectors_Vector{
			Vector: &pb.Vector{
				Data: embedding,
			},
		},
	}

	// payload
	ret := make(map[string]*pb.Value)
	ret["combined"] = &pb.Value{Kind: &pb.Value_StringValue{StringValue: combined}}
	ret["raw"] = &pb.Value{Kind: &pb.Value_StringValue{StringValue: getRawVector(combined)}}
	point.Payload = ret
	return point
}

func md5str(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func convertToFloat32(embedding []float64) []float32 {
	var ret []float32
	for _, v := range embedding {
		ret = append(ret, float32(v))
	}
	return ret
}

func getRawVector(combined string) string {
	// note value may contain duplicate words to another sample data
	// and will reduce the accuracy of the embedding
	indexNote1 := strings.Index(combined, "note1:")
	if indexNote1 > 0 {
		combined = combined[:indexNote1]
	}

	raw := strings.ReplaceAll(combined, "stock_no:", "")
	raw = strings.ReplaceAll(raw, "id_lelang:", "")
	raw = strings.ReplaceAll(raw, "cabang:", "")
	raw = strings.ReplaceAll(raw, "tanggal:", "")
	raw = strings.ReplaceAll(raw, "bulan:", "")
	raw = strings.ReplaceAll(raw, "jalur:", "")
	raw = strings.ReplaceAll(raw, "lot:", "")
	raw = strings.ReplaceAll(raw, "seller_no:", "")
	raw = strings.ReplaceAll(raw, "seller_name:", "")
	raw = strings.ReplaceAll(raw, "seller_kategori:", "")
	raw = strings.ReplaceAll(raw, "npwp_penjual:", "")
	raw = strings.ReplaceAll(raw, "alamat_penjual:", "")
	raw = strings.ReplaceAll(raw, "nomor_kontrak:", "")
	raw = strings.ReplaceAll(raw, "nama:", "")
	raw = strings.ReplaceAll(raw, "plat_no:", "")
	raw = strings.ReplaceAll(raw, "pabrikan:", "")
	raw = strings.ReplaceAll(raw, "model:", "")
	raw = strings.ReplaceAll(raw, "type:", "")
	raw = strings.ReplaceAll(raw, "tahun:", "")
	raw = strings.ReplaceAll(raw, "transmisi:", "")
	raw = strings.ReplaceAll(raw, "warna:", "")
	raw = strings.ReplaceAll(raw, "harga_awal:", "")
	raw = strings.ReplaceAll(raw, "harga_terbentuk:", "")
	raw = strings.ReplaceAll(raw, "dpp:", "")
	raw = strings.ReplaceAll(raw, "ppn_1,1%:", "")
	raw = strings.ReplaceAll(raw, "status:", "")
	raw = strings.ReplaceAll(raw, "segment:", "")
	raw = strings.ReplaceAll(raw, "kapasitas_mesin:", "")
	raw = strings.ReplaceAll(raw, "tipe_bahan_bakar:", "")
	raw = strings.ReplaceAll(raw, "odometer:", "")
	raw = strings.ReplaceAll(raw, " grade:", "")
	raw = strings.ReplaceAll(raw, "no_mesin:", "")
	raw = strings.ReplaceAll(raw, "no_rangka:", "")
	raw = strings.ReplaceAll(raw, "status_bpkb:", "")
	raw = strings.ReplaceAll(raw, "no_bpkb:", "")
	raw = strings.ReplaceAll(raw, "nama_bpkb:", "")
	raw = strings.ReplaceAll(raw, "status_stnk:", "")
	raw = strings.ReplaceAll(raw, "no_stnk:", "")
	raw = strings.ReplaceAll(raw, "nama_stnk:", "")
	raw = strings.ReplaceAll(raw, "stnk_exp_date:", "")
	raw = strings.ReplaceAll(raw, "faktur:", "")
	raw = strings.ReplaceAll(raw, "kwitansi_blank:", "")
	raw = strings.ReplaceAll(raw, "fc_ktp:", "")
	raw = strings.ReplaceAll(raw, "form_a:", "")
	raw = strings.ReplaceAll(raw, "status_keur:", "")
	raw = strings.ReplaceAll(raw, "masa_berlaku_keur:", "")
	raw = strings.ReplaceAll(raw, "nopol_nipl:", "")
	raw = strings.ReplaceAll(raw, "no_pembeli:", "")
	raw = strings.ReplaceAll(raw, "eksterior_grade:", "")
	raw = strings.ReplaceAll(raw, "interior_grade:", "")
	raw = strings.ReplaceAll(raw, "mesin_grade:", "")
	raw = strings.ReplaceAll(raw, "note1:", "")
	raw = strings.ReplaceAll(raw, "note2:", "")
	raw = strings.ReplaceAll(raw, "rongsokan:", "")
	raw = strings.ReplaceAll(raw, "time_closed:", "")
	raw = strings.ReplaceAll(raw, "va_payment:", "")
	raw = strings.ReplaceAll(raw, "nomor_telepon_penjual:", "")
	raw = strings.ReplaceAll(raw, "nama_pembeli:", "")
	raw = strings.ReplaceAll(raw, "alamat_pembeli:", "")
	raw = strings.ReplaceAll(raw, "nomor_handphone:", "")
	raw = strings.ReplaceAll(raw, "no_ktp_atau_passport:", "")
	raw = strings.ReplaceAll(raw, "npwp_pembeli:", "")
	raw = strings.ReplaceAll(raw, "eksterior_grade:", "")
	raw = strings.ReplaceAll(raw, "interior_grade:", "")
	raw = strings.ReplaceAll(raw, "mesin_grade:", "")
	raw = strings.ReplaceAll(raw, " ; ", "")
	raw = strings.ReplaceAll(raw, "-; ", "")
	raw = strings.ReplaceAll(raw, "null; ", "")
	raw = strings.ReplaceAll(raw, "; ", "")

	return raw
}
