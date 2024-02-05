package elasticsearch

import (
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"strconv"
	"strings"
)

type ESClient struct {
	client *elasticsearch.Client
	index  string
}

type ESSearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		}
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index   string   `json:"_index"`
			ID      string   `json:"_id"`
			Score   float64  `json:"_score"`
			Ignored []string `json:"_ignored"`
			Source  struct {
				Combined  string    `json:"combined"`
				Raw       string    `json:"raw"`
				Embedding []float64 `json:"embedding"`
			} `json:"_source"`
		} `json:"hits"`
	}
}

func NewElasticsearch() *ESClient {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	return &ESClient{
		client: es,
	}
}

func (es *ESClient) GetClient() *elasticsearch.Client {
	return es.client
}

func (es *ESClient) SetIndex(index string) {
	es.index = index
}

func (es *ESClient) GetIndex() string {
	return es.index
}

func (es *ESClient) CreateIndex() error {
	mapping := `
    {
      "settings": {
        "number_of_shards": 1
      },
      "mappings": {
        "properties": {
          "embedding": {
            "type": "dense_vector",
			"dims": 1536,
			"index": "true",
			"similarity": "cosine"
          },
		  "combined": "text",
		  "raw": "text",
        }
      }
    }`

	_, err := es.client.Indices.Create(es.index, es.client.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return err
	}

	return nil
}

func (es *ESClient) DeleteIndex() error {
	_, err := es.client.Indices.Delete([]string{es.index})
	if err != nil {
		return err
	}

	return nil
}

func (es *ESClient) IndexDocument(document map[string]interface{}) error {
	documentByte, err := json.Marshal(document)
	if err != nil {
		return err
	}

	_, err = es.client.Index(es.index, bytes.NewBuffer(documentByte))
	if err != nil {
		return err
	}

	return nil
}

func (es *ESClient) VectorSearch(vector []float64) (*ESSearchResponse, error) {
	var strArr []string
	for _, v := range vector {
		strArr = append(strArr, strconv.FormatFloat(v, 'f', -1, 64))
	}

	log.Println("vector: ", strings.Join(strArr, ", "))

	//query := `
	//{
	// "query": {
	//	"script_score": {
	//	  "query": {
	//		"match_all": {}
	//	  },
	//	  "script": {
	//		"source": "cosineSimilarity(params.query_vector, 'embedding') + 1.0",
	//		"params": {
	//		  "query_vector": [` + strings.Join(strArr, ", ") + `]
	//		}
	//	  }
	//	}
	// }
	//}`

	query := `
	{
	 "knn": {	
		"field": "embedding",
		"query_vector": [` + strings.Join(strArr, ", ") + `],
		"k": 3,
		"num_candidates": 3
	 },
	 "rescore": {
		"window_size": 3,
		"query": {
			"rescore_query": {
				"script_score": {
					"query": {
						"match_all": {}
					},
					"script": {
						"source": "cosineSimilarity(params.query_vector, 'embedding') + 1.0",
						"params": {
							"query_vector": [` + strings.Join(strArr, ", ") + `]
						}
					}
				}
			}
		}
	 }
	}`

	res, err := es.client.Search(
		es.client.Search.WithIndex(es.index),
		es.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, err
	}

	var result *ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (es *ESClient) IndexSearch(question string) (*ESSearchResponse, error) {
	query := `
	{
	 "_source": {
			"excludes": ["embedding"]	
		},
		"query": {
			"query_string": {
				"fields": [
					"raw"
				],
				"query": "` + question + `",
				"minimum_should_match": 1
			}
		}
	}`
	res, err := es.client.Search(
		es.client.Search.WithIndex(es.index),
		es.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, err
	}

	var result *ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (es *ESClient) HybridSearch(vector []float64, question string) (*ESSearchResponse, error) {
	var strArr []string
	for _, v := range vector {
		strArr = append(strArr, strconv.FormatFloat(v, 'f', -1, 64))
	}

	log.Println("vector: ", strings.Join(strArr, ", "))

	query := `
	{
	 "knn": {	
		"field": "embedding",
		"query_vector": [` + strings.Join(strArr, ", ") + `],
		"k": 3,
		"num_candidates": 3,
		"boost": 0.1
	 },
	"query": {
		"query_string": {
			"fields": [
				"raw"
			],
			"query": "` + question + `",
			"minimum_should_match": 1,
			"boost": 0.9
		}
	},
	"size": 3
	}`

	res, err := es.client.Search(
		es.client.Search.WithIndex(es.index),
		es.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, err
	}

	var result *ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
