package handler

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allowedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
}

func saveUploadedImage(c *gin.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := allowedImageExtensions[ext]; !ok {
		return "", fmt.Errorf("unsupported image type: %s", ext)
	}

	fileID := uuid.New().String() + ext
	dir := filepath.Join("uploads", folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	fullPath := filepath.Join(dir, fileID)
	if err := c.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return "", fmt.Errorf("failed to save uploaded file: %w", err)
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	publicPath := "/" + filepath.ToSlash(filepath.Join("uploads", folder, fileID))
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, publicPath), nil
}
