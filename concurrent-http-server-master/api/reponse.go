package api

type uploadResponseBody struct {
	FileId string `json:"file_id"`
}

type failureResponseBody struct {
	Error string `json:"error"`
}
