package controllers

import (
	"net/http"
	"strconv"

	"cky.backend/services"
	"github.com/gin-gonic/gin"
)

type FileController struct {
	fileService services.FileService
}

func NewFileController() *FileController {
	return &FileController{
		fileService: services.NewFileService(),
	}
}

func (c *FileController) UploadChunk(ctx *gin.Context) {
	// 取得分塊資訊，如: 文件ID，分塊編號等
	fileID := ctx.PostForm("file_id")
	chunkNumberStr := ctx.PostForm("chunkIndex")
	chunkNumber, err := strconv.Atoi(chunkNumberStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk number"})
		return
	}

	filePart, err := ctx.FormFile("chunk")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// 將分塊儲存到服務端
	if err := c.fileService.SaveChunk(ctx, fileID, chunkNumber, filePart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "Chunk uploaded successfully"})
}

func (c *FileController) CompleteUpload(ctx *gin.Context) {
	var body UploadCompleteParams
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.fileService.MergeChunks(body.FileId, body.TotalChunks)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "uploaded successfully"})
}

type UploadCompleteParams struct {
	FileId      string `json:"fileId"`
	TotalChunks int    `json:"totalChunks"`
}
