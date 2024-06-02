package main

import (
	"cky.backend/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	fileController := controllers.NewFileController()
	r.Use(cors.Default())
	r.POST("/upload", fileController.UploadChunk)
	r.POST("/upload/complete", fileController.CompleteUpload)

	r.Run("localhost:8080")
}
