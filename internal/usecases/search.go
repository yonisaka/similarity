package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/internal/entities"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	embeddingURL = "https://api.openai.com/v1/embeddings"
	openaiModel  = "gpt-3.5-turbo"
	roleUser     = "user"
	filepath     = "../data/sample_data.json"
	introduction = "Use the below sample data to answer the subsequent question. If the answer cannot be found in the data source, write \"I could not find an answer.\""
)

func (u *searchUsecase) Search(ctx context.Context, query string) (string, error) {
	records, err := u.LoadDataSources(filepath)
	if err != nil {
		return "", err
	}

	recordsAndRelatedness, err := u.StringsRankedByRelatedness(query, records, 10)
	if err != nil {
		return "", err
	}

	// Ask a question using the top N strings
	answer, err := u.Ask(ctx, query, recordsAndRelatedness, 1000) // Adjust the token budget as needed
	if err != nil {
		return "", err
	}

	return answer, nil
}

func (u *searchUsecase) LoadDataSources(filepath string) ([]entities.Record, error) {
	// Load and parse JSON data
	file, err := os.Open(filepath) // Adjust the path to your JSON file
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []entities.Record
	if err := json.NewDecoder(file).Decode(&records); err != nil {
		return nil, err
	}

	return records, nil
}

func (u *searchUsecase) CosineSimilarity(vecA, vecB []float64) float64 {
	var dotProduct float64
	var normA, normB float64
	for i, val := range vecA {
		dotProduct += val * vecB[i]
		normA += val * val
		normB += vecB[i] * vecB[i]
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (u *searchUsecase) EmbeddingQuery(query string) ([]float64, error) {
	// Construct the request body
	requestBody, err := json.Marshal(map[string]string{
		"input": query,
	})
	if err != nil {
		log.Fatalf("Error occurred while marshaling. %s", err)
	}

	// Create an HTTP request
	req, err := http.NewRequest("POST", embeddingURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", os.Getenv("OPENAI_API_KEY"))

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
	var embeddingResponse entities.EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		return nil, err
	}

	// Return the embedding
	if len(embeddingResponse.Data) > 0 {
		return embeddingResponse.Data[0].Embedding, nil
	}

	return nil, nil
}

// StringsRankedByRelatedness finds strings ranked by their relatedness to a query.
func (u *searchUsecase) StringsRankedByRelatedness(query string, records []entities.Record, topN int) ([]entities.StringAndRelatedness, error) {
	queryEmbedding, err := u.EmbeddingQuery(query)
	if err != nil {
		return nil, err
	}

	results := make([]entities.StringAndRelatedness, 0, len(records))
	for _, record := range records {
		relatedness := u.CosineSimilarity(queryEmbedding, record.Embedding)
		results = append(results, entities.StringAndRelatedness{Text: record.Combined, Relatedness: relatedness})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Relatedness > results[j].Relatedness
	})

	if topN > len(results) {
		topN = len(results)
	}

	return results[:topN], nil
}

// NumTokens approximates the number of tokens in a string.
func (u *searchUsecase) NumTokens(text string) int {
	// Simple tokenization by splitting on spaces and punctuation
	// Adjust this as needed for a more accurate count
	return len(strings.Fields(text))
}

// QueryMessage builds a message with relevant texts from the data.
func (u *searchUsecase) QueryMessage(query string, records []entities.StringAndRelatedness, tokenBudget int) string {
	question := "\n\nQuestion: " + query
	message := introduction
	for _, record := range records {
		nextPrompt := "\n\nPrompt section:\n\"\"\"\n" + record.Text + "\n\"\"\""
		if u.NumTokens(message+nextPrompt+question) > tokenBudget {
			break
		} else {
			message += nextPrompt
		}
	}
	return message + question
}

// Ask answers a query using GPT and a slice of relevant texts and embeddings.
func (u *searchUsecase) Ask(ctx context.Context, query string, records []entities.StringAndRelatedness, tokenBudget int) (string, error) {
	message := u.QueryMessage(query, records, tokenBudget)

	resp, err := u.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openaiModel, // Adjust the model as needed
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    roleUser,
				Content: message,
			},
		},
		Temperature: 0,
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
