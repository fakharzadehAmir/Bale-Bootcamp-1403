package main

import (
	"concurrent/api"
	httpserver "concurrent/internal/http_server"
	"concurrent/pkg"
	"github.com/sirupsen/logrus"
	"net/http"
)

// @title Upload/Download File
// @version 1.0
// @description Upload/Download File Project Bale-Bootcamp-403
// @host localhost:8080
// @BasePath /
func main() {
	//	Logger
	logger := logrus.New()
	//	FileDB
	fileDB := pkg.NewFileTable()
	logger.Info("fileDB created successfully")
	//	FileServer
	fileServer := httpserver.NewFileServer(fileDB, logger)
	logger.Info("file server created successfully")

	fileModule := api.NewFileModule(fileServer, logger)
	logger.Info("file module created successfully")

	routes := fileModule.GetRoutes()

	for _, route := range routes {
		http.HandleFunc(route.Path, route.HandlerFunc)
	}

	logger.Info("is running:")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
