package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CalculateFileHash calculates the MD5 hash of a file
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// RemoveFile safely removes a file if it exists
func RemoveFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// GetFileExtension returns the extension of a file (including the dot)
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// GetFileNameWithoutExtension returns the filename without extension
func GetFileNameWithoutExtension(filename string) string {
	extension := filepath.Ext(filename)
	return strings.TrimSuffix(filename, extension)
}

// CopyFile copies a file from source to destination
func CopyFile(sourcePath, destPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Ensure we write all data to disk
	err = destFile.Sync()
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	return os.Chmod(destPath, sourceInfo.Mode())
}

// MoveFile moves a file from source to destination
func MoveFile(sourcePath, destPath string) error {
	// First try with os.Rename (which might fail across different filesystems)
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}

	// If rename fails, try copy and delete
	err = CopyFile(sourcePath, destPath)
	if err != nil {
		return err
	}

	return os.Remove(sourcePath)
}

// IsFileExists checks if a file exists
func IsFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirExists checks if a directory exists
func IsDirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateTempFile creates a temporary file with the given prefix and content
func CreateTempFile(prefix string, content []byte) (string, error) {
	tempFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if content != nil {
		_, err = tempFile.Write(content)
		if err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	}

	return tempFile.Name(), nil
}

// GetMimeType detects the MIME type of a file based on its extension
// This is a simple implementation and not as accurate as using http.DetectContentType
func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".wav":
		return "audio/wav"
	case ".avi":
		return "video/x-msvideo"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
