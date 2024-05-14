package api

type uploadRequestBody struct {
	File string `json:"file"`
}

type downloadRequestBody struct {
	FileId string `json:"file_id"`
}
