package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/internal/entities"
	"github.com/yonisaka/similarity/internal/entities/repository"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	embeddingURL         = "https://api.openai.com/v1/embeddings"
	openaiModel          = "gpt-3.5-turbo"
	textEmbeddingAda     = "text-embedding-ada-002"
	textEmbedding3Small  = "text-embedding-3-small"
	roleUser             = "user"
	filepath             = "../data/sample_data.json"
	similarityQdrant     = "qdrant"
	similarityPostgresql = "postgresql"
	topN                 = 5
	tokenBudget          = 1000
	introduction         = "Use the below sample data to answer the subsequent question. If the answer cannot be found in the data source, write \"I could not find an answer.\""
)

func (u *searchUsecase) Search(ctx context.Context, query string) (string, error) {
	var recordsAndRelatedness []entities.StringAndRelatedness
	var err error

	if os.Getenv("SIMILARITY_METHOD") == similarityQdrant {
		recordsAndRelatedness, err = u.QdrantSearch(ctx, query)
		if err != nil {
			return "", err
		}
	} else if os.Getenv("SIMILARITY_METHOD") == similarityPostgresql {
		records, err := u.embeddingRepo.ListEmbeddingByScope(ctx, "sample_lelang.csv")
		if err != nil {
			return "", err
		}

		recordsAndRelatedness, err = u.StringsRankedByRelatedness(query, records, topN)
		if err != nil {
			return "", err
		}
	}

	// Ask a question using the top N strings
	answer, err := u.Ask(ctx, query, recordsAndRelatedness, tokenBudget) // Adjust the token budget as needed
	if err != nil {
		return "", err
	}

	return answer, nil
}

func (u *searchUsecase) LoadJSONDataSources(filepath string) ([]repository.Embedding, error) {
	// Load and parse JSON data
	file, err := os.Open(filepath) // Adjust the path to your JSON file
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []repository.Embedding
	if err := json.NewDecoder(file).Decode(&records); err != nil {
		return nil, err
	}

	return records, nil
}

func (u *searchUsecase) EmbeddingQuery(query string) ([]float64, error) {
	// Construct the request body
	requestBody, err := json.Marshal(map[string]string{
		"input": query,
		"model": os.Getenv("OPENAI_EMBEDDING_MODEL"),
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
		return nil, err
	}

	// Return the embedding
	if len(embeddingResponse.Data) > 0 {
		return embeddingResponse.Data[0].Embedding, nil
	}

	var errorResponse *entities.ErrorResponse
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return nil, err
	}

	if errorResponse != nil {
		return nil, errors.New(errorResponse.Error.Message)
	}

	return nil, nil
}

// StringsRankedByRelatedness finds strings ranked by their relatedness to a query.
func (u *searchUsecase) StringsRankedByRelatedness(query string, records []repository.Embedding, topN int) ([]entities.StringAndRelatedness, error) {
	queryEmbedding, err := u.EmbeddingQuery(query)
	if err != nil {
		return nil, err
	}

	results := make([]entities.StringAndRelatedness, 0, len(records))
	for _, record := range records {
		relatedness, err := cosineSimilarity(queryEmbedding, record.Embedding)
		if err != nil {
			return nil, err
		}
		results = append(results, entities.StringAndRelatedness{
			ID:          record.ID,
			Text:        record.Combined,
			Relatedness: relatedness,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Relatedness > results[j].Relatedness
	})

	if topN > len(results) {
		topN = len(results)
	}

	for _, result := range results[:topN] {
		fmt.Println(fmt.Sprintf("record id: %d relatedness: %f", result.ID, result.Relatedness))
	}

	return results[:topN], nil
}

func (u *searchUsecase) QdrantSearch(ctx context.Context, query string) ([]entities.StringAndRelatedness, error) {
	queryEmbedding, err := u.EmbeddingQuery(query)
	if err != nil {
		return nil, err
	}

	points, err := u.qdrantClient.Search(ctx, convertToFloat32(queryEmbedding))
	if err != nil {
		return nil, err
	}

	if len(points) == 0 {
		return nil, errors.New("no result found")
	}

	results := make([]entities.StringAndRelatedness, 0, len(points))
	// best score from ascending order
	for _, point := range points {
		results = append(results, entities.StringAndRelatedness{
			QdrantID:    point.Id.GetUuid(),
			Text:        point.Payload["combined"].GetStringValue(),
			Relatedness: float64(point.Score),
		})

		fmt.Println(fmt.Sprintf("record id: %s relatedness: %f", point.Id.GetUuid(), point.Score))
	}

	// using scroll
	// to get specific record by user prompt input
	// must match one of the word in the query
	// if scroll result has been existed in results, skip it
	usingScroll, err := strconv.ParseBool(os.Getenv("QDRANT_SCROLL"))
	if err != nil {
		return nil, err
	}

	if usingScroll {
		//scrolls, err := u.qdrantClient.Scroll(ctx, query)
		scrolls, err := u.qdrantClient.MultiScroll(ctx, query)
		if err != nil {
			return nil, err
		}

		for _, scroll := range scrolls {
			existID := false
			for _, result := range results {
				if scroll.Id.GetUuid() == result.QdrantID {
					existID = true
					break
				}
			}

			if existID {
				continue
			}

			results = append(results, entities.StringAndRelatedness{
				QdrantID: scroll.Id.GetUuid(),
				Text:     scroll.Payload["combined"].GetStringValue(),
			})

			fmt.Println(fmt.Sprintf("record id: %s ", scroll.Id.GetUuid()))
		}
	}

	return results, nil
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
