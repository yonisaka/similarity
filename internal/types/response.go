package types

type Http struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type SearchResponse struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
