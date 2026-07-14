package handler

// envelope is the standard JSON response wrapper.
// Using a consistent shape lets clients write one parser.
type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *errBody    `json:"error,omitempty"`
}

type errBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
