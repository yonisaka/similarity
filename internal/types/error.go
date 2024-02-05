package types

// ErrorResponse represents the expected structure of the response from your embedding API
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
