package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FileService struct {
}

func NewFileService() FileService {
	return FileService{}
}

func (s *FileService) SaveChunk(ctx *gin.Context, fileID string, chunkNumber int, filePart *multipart.FileHeader) error {
	// 這裡設置分塊的儲存位置
	chunkDir := filepath.Join("uploads_temp", fileID)
	if err := os.MkdirAll(chunkDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	dst := filepath.Join(chunkDir, fmt.Sprintf("%d.part", chunkNumber))
	if err := ctx.SaveUploadedFile(filePart, dst); err != nil {
		return fmt.Errorf("failed to save file part: %v", err)
	}

	return nil
}

func (s *FileService) MergeChunks(fileID string, totalChunks int) error {
	chunkTempDir := filepath.Join("uploads_temp", fileID)
	chunkDir := filepath.Join("uploads")

	if err := os.MkdirAll(chunkDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	outputFile := filepath.Join("uploads", fileID)

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer out.Close()

	for i := 1; i <= totalChunks; i++ {
		chunkFile := filepath.Join(chunkTempDir, fmt.Sprintf("%d.part", i))
		in, err := os.Open(chunkFile)
		if err != nil {
			return fmt.Errorf("failed to open chunk file: %v", err)
		}
		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			return fmt.Errorf("failed to write chunk file: %v", err)
		}
	}

	if err := os.RemoveAll(chunkTempDir); err != nil {
		return fmt.Errorf("failed to delete temp folder: %v", err)
	}

	return nil
}
