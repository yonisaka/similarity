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
)

const (
	maxTokens = 2000
)

var errGetEmbedding = errors.New("error getting embedding")

func (u *importUsecase) Import(ctx context.Context, filename string, offset int) error {
	combined, rawVectors, err := u.ReadCSV(filename)
	if err != nil {
		return err
	}

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

		log.Println(fmt.Sprintf("used tokens: %d", tokens))
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
			if i == 0 {
				combine = fmt.Sprintf("%s: %s", headers[i], field)
				rawVector = field
			} else {
				combine = fmt.Sprintf("%s; %s: %s", combine, headers[i], field)
				rawVector = fmt.Sprintf("%s; %s", rawVector, field)
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
		"model": textEmbedding3Small,
	})
	if err != nil {
		log.Fatalf("Error occurred while marshaling. %s", err)
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
		log.Fatalf("Error occurred while making HTTP request. %s", err)
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
