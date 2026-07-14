package handler

// envelope is the standard JSON response wrapper.
// It helps to have a consistent response shape so the frontend developers don't get mad at us!
type envelope struct {
	Success bool        `json:"success"`        // True if the request worked, false if it failed
	Data    interface{} `json:"data,omitempty"` // interface{} means it can hold any type of data (e.g., User or list of Users). "omitempty" hides this field if it is empty!
	Error   *errBody    `json:"error,omitempty"` // Pointer to errBody so it can be nil. If it's nil, "omitempty" hides it from the JSON.
}

// errBody is what we send back when something goes wrong.
type errBody struct {
	Code    string `json:"code"`    // A short string code representing the error (like INVALID_ID)
	Message string `json:"message"` // A human-readable description of what went wrong
}
