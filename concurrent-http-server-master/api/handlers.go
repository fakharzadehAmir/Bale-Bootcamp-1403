package api

import (
	httpserver "concurrent/internal/http_server"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type FileModule struct {
	// Configs
	// Logrus
	logger     *logrus.Logger
	fileServer *httpserver.FileServer
}

func NewFileModule(fileServer *httpserver.FileServer, logger *logrus.Logger) IModule {
	return &FileModule{
		fileServer: fileServer,
		logger:     logger,
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(failureResponseBody{Error: message})
}

func (fm *FileModule) GetRoutes() []Route {
	return []Route{
		*NewRoute("/downloadFile", http.MethodGet, fm.downloadFileHandler),
		*NewRoute("/uploadFile", http.MethodPost, fm.uploadFileHandler),
	}
}

func downloadFileByURL(url string) (io.ReadCloser, string, error) {
	responseDownload, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	// consider file name
	return responseDownload.Body, "fileName.pdf", nil
}

func (fm *FileModule) uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		var reqBody = &uploadRequestBody{}
		err := json.NewDecoder(r.Body).Decode(reqBody)
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusBadRequest, "Invalid JSON data")
			return
		}

		file, fileName, err := downloadFileByURL(reqBody.File)
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer file.Close()

		byteData, err := io.ReadAll(file)
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusInternalServerError, "Failed to read the file data")
			return
		}

		fileId, err := fm.fileServer.WriteToFile(byteData, fileName)
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusInternalServerError, "Failed to write file data to the server")
			return
		}

		response := uploadResponseBody{FileId: fileId}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	} else if contentType == "multipart/form-data; boundary=X-INSOMNIA-BOUNDARY" {
		if err := r.ParseMultipartForm(4 << 20); err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusBadRequest, "Error parsing multipart form")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusBadRequest, "Failed to get the file")
			return
		}
		fileName := header.Filename
		defer file.Close()

		byteData, err := io.ReadAll(file)

		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusInternalServerError, "Failed to read the file data")
			return
		}

		fileId, err := fm.fileServer.WriteToFile(byteData, fileName)
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusInternalServerError, "Failed to write file data to the server")
			return
		}

		response := uploadResponseBody{FileId: fileId}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	} else {
		fm.logger.WithError(errors.New("invalid content-type")).Warn("Error:")
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type")
		return
	}

}

func (fm *FileModule) downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	var reqBody = &downloadRequestBody{}
	if contentType == "application/json" {
		err := json.NewDecoder(r.Body).Decode(reqBody)
		if err != nil {
			fm.logger.WithError(err).Warn("Error:")
			respondWithError(w, http.StatusBadRequest, "Invalid JSON data")
			return
		}

	} else {
		fm.logger.WithError(errors.New("invalid content-type")).Warn("Error:")
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type")
		return
	}

	file, fileName, err := fm.fileServer.ReadFromFile(reqBody.FileId)
	if err != nil {
		fm.logger.WithError(err).Warn("Error:")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.WriteHeader(http.StatusOK)
	w.Write(file)
}
