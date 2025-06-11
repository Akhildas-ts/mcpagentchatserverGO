package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SplitIntoChunks splits content into chunks of approximately the specified size
func SplitIntoChunks(content string, chunkSize int) []string {
	lines := strings.Split(content, "\n")
	chunks := []string{}
	currentChunk := ""
	currentSize := 0

	for _, line := range lines {
		lineSize := len(line)
		if currentSize+lineSize > chunkSize && currentSize > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = line
			currentSize = lineSize
		} else {
			if currentSize > 0 {
				currentChunk += "\n"
			}
			currentChunk += line
			currentSize += lineSize + 1 // +1 for newline
		}
	}

	if currentSize > 0 {
		chunks = append(chunks, currentChunk)
	}

	fmt.Printf("Split into %d chunks\n", len(chunks))

	return chunks
}

// GetLanguageFromExtension returns the programming language based on file extension
func GetLanguageFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".go":
		return "Go"
	case ".js", ".jsx":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".java":
		return "Java"
	case ".c", ".cpp", ".h", ".hpp":
		return "C/C++"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".cs":
		return "C#"
	case ".html":
		return "HTML"
	case ".css":
		return "CSS"
	default:
		return "Unknown"
	}
}

// IsBinaryFile checks if a file is binary based on its extension
func IsBinaryFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	binaryExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".ico", ".svg",
		".zip", ".tar", ".gz", ".rar", ".7z",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".mp3", ".mp4", ".wav", ".avi", ".mov",
		".so", ".dll", ".exe", ".bin",
	}

	for _, binaryExt := range binaryExtensions {
		if ext == binaryExt {
			return true
		}
	}

	// Check filename for Git internal files
	if strings.Contains(filename, ".git/") ||
		strings.HasPrefix(filename, ".git") ||
		filename == "DIRC" ||
		strings.Contains(filename, "index.lock") {
		return true
	}

	return false
}

// ContainsBinaryData checks if content contains binary data
func ContainsBinaryData(content []byte) bool {
	// Check first few bytes for NULL or other binary indicators
	for i := 0; i < min(len(content), 1000); i++ {
		if content[i] == 0 || content[i] > 127 {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}