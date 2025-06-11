package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"mcpserver/internal/models"
	"mcpserver/internal/storage"
	"mcpserver/pkg/git"
	"mcpserver/pkg/utils"
)

type RepoIndexerService struct {
	pineconeStore *storage.PineconeStore
	openaiClient  *storage.OpenAIClient
}

func NewRepoIndexerService(pineconeStore *storage.PineconeStore, openaiClient *storage.OpenAIClient) *RepoIndexerService {
	return &RepoIndexerService{
		pineconeStore: pineconeStore,
		openaiClient:  openaiClient,
	}
}

func (ri *RepoIndexerService) IndexRepository(repoURL, branch string) error {
	// Extract repository name from URL
	parts := strings.Split(repoURL, "/")
	repoName := parts[len(parts)-1]
	if strings.HasSuffix(repoName, ".git") {
		repoName = repoName[:len(repoName)-4]
	}

	fmt.Printf("Indexing repository: %s, branch: %s\n", repoURL, branch)

	// Create temporary directory for cloning
	tempDir, err := ioutil.TempDir("", "repo-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Created temp directory: %s\n", tempDir)

	// Clone repository using git package
	if err := git.CloneRepository(repoURL, tempDir, branch); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("Cloned repository to: %s\n", tempDir)

	// Process repository files
	return ri.processDirectory(tempDir, repoURL, branch)
}

func (ri *RepoIndexerService) processDirectory(dir, repoURL, branch string) error {
	fileCount := 0
	skippedCount := 0
	processedCount := 0

	baseDirLen := len(dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}

		fileCount++
		relPath := path
		if len(path) > baseDirLen {
			relPath = path[baseDirLen+1:]
		}

		// Skip directories and hidden files
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				if strings.Contains(path, ".git") {
					fmt.Printf("Skipping .git directory: %s\n", relPath)
					skippedCount++
					return filepath.SkipDir // Skip .git directories entirely
				}
				fmt.Printf("Skipping hidden directory: %s\n", relPath)
				skippedCount++
				return nil
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			fmt.Printf("Skipping hidden file: %s\n", relPath)
			skippedCount++
			return nil
		}

		// Skip binary files based on extension
		if utils.IsBinaryFile(path) {
			fmt.Printf("Skipping binary file: %s\n", relPath)
			skippedCount++
			return nil
		}

		// Read file content
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", relPath, err)
			skippedCount++
			return nil // Skip files we can't read
		}

		// Skip large files and binary content
		if len(content) > 100000 {
			fmt.Printf("Skipping large file: %s (%d bytes)\n", relPath, len(content))
			skippedCount++
			return nil
		}

		if utils.ContainsBinaryData(content) {
			fmt.Printf("Skipping file with binary data: %s\n", relPath)
			skippedCount++
			return nil
		}

		// Process file content
		fmt.Printf("Processing file: %s\n", relPath)
		if err := ri.processFile(string(content), path, repoURL, branch); err != nil {
			fmt.Printf("Error processing file %s: %v\n", path, err)
			skippedCount++
			return nil // Continue with other files even if one fails
		}

		processedCount++
		if processedCount%10 == 0 {
			fmt.Printf("Processed %d files so far...\n", processedCount)
		}

		return nil
	})

	fmt.Printf("Directory processing complete. Total files: %d, Skipped: %d, Processed: %d\n",
		fileCount, skippedCount, processedCount)

	return err
}

func (ri *RepoIndexerService) processFile(content, filePath, repoURL, branch string) error {
	// Extract repository name from URL
	parts := strings.Split(repoURL, "/")
	repoOwner := parts[len(parts)-2]
	repoName := parts[len(parts)-1]
	if strings.HasSuffix(repoName, ".git") {
		repoName = repoName[:len(repoName)-4]
	}
	repository := fmt.Sprintf("%s/%s", repoOwner, repoName)

	// Get relative file path - extract the path after the temp directory
	relPath := filePath
	tempDirMarker := "/repo-"
	if idx := strings.LastIndex(filePath, tempDirMarker); idx != -1 {
		// Find the next path separator after the temp dir name
		nextSlash := strings.Index(filePath[idx+1:], "/")
		if nextSlash != -1 {
			relPath = filePath[idx+nextSlash+2:] // +2 to account for the slash and the idx+1 offset
		}
	}

	fmt.Printf("Processing file %s with relative path %s\n", filePath, relPath)

	// Determine language from file extension
	language := utils.GetLanguageFromExtension(filepath.Ext(filePath))

	// Split content into chunks of approximately 1000 tokens
	chunks := utils.SplitIntoChunks(content, 1000)
	fmt.Printf("Split into %d chunks\n", len(chunks))

	// Process each chunk
	for i, chunk := range chunks {
		// Get embedding for the chunk
		embedding, err := ri.openaiClient.GetEmbedding(chunk)
		if err != nil {
			return fmt.Errorf("failed to get embedding: %w", err)
		}

		// Create code chunk
		codeChunk := models.CodeChunk{
			Content:    chunk,
			FilePath:   relPath,
			Repository: repository,
			Branch:     branch,
			Language:   language,
			Embedding:  embedding,
		}

		fmt.Printf("Storing chunk for %s, repository %s\n", relPath, repository)

		// Store in vector database
		if err := ri.pineconeStore.Store(codeChunk); err != nil {
			return fmt.Errorf("failed to store chunk: %w", err)
		}

		if i == 0 || i%10 == 0 {
			fmt.Printf("Indexed chunk %d for file: %s\n", i, relPath)
		}
	}

	return nil
}