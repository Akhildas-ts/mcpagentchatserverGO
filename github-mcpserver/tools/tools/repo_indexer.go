package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClientInterface defines the interface for OpenAI client operations
type OpenAIClientInterface interface {
	CreateEmbeddings(ctx context.Context, request openai.EmbeddingRequestConverter) (openai.EmbeddingResponse, error)
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// OpenAIClientWrapper wraps the OpenAI client to match our interface
type OpenAIClientWrapper struct {
	client *openai.Client
}

func NewOpenAIClientWrapper(apiKey string) *OpenAIClientWrapper {
	return &OpenAIClientWrapper{
		client: openai.NewClient(apiKey),
	}
}

func (w *OpenAIClientWrapper) CreateEmbeddings(ctx context.Context, request openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	return w.client.CreateEmbeddings(ctx, openai.EmbeddingRequestConverter(request))
}

func (w *OpenAIClientWrapper) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return w.client.CreateChatCompletion(ctx, request)
}

// Compile-time check to ensure openai.Client implements OpenAIClientInterface
var _ OpenAIClientInterface = (*openai.Client)(nil)

// RepoIndexer handles repository indexing operations
type RepoIndexer struct {
	vectorStore  VectorStoreInterface
	openAIClient OpenAIClientInterface
}

// NewRepoIndexer creates a new RepoIndexer instance
func NewRepoIndexer(pineconeKey, pineconeEnv, indexName, pineconeHost, openAIKey string) (*RepoIndexer, error) {
	if openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	vectorStore, err := NewVectorStore(pineconeKey, pineconeEnv, indexName, pineconeHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	openAIClient := openai.NewClient(openAIKey)

	return &RepoIndexer{
		vectorStore:  vectorStore,
		openAIClient: openAIClient,
	}, nil
}

// IndexRepository handles the full process of cloning and indexing a repository
func (r *RepoIndexer) IndexRepository(repoURL, branch string) error {
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

	// Clone repository
	cmd := exec.Command("git", "clone", repoURL, tempDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("Cloned repository to: %s\n", tempDir)

	// Checkout specific branch if specified
	if branch != "" && branch != "main" && branch != "master" {
		cmd = exec.Command("git", "checkout", branch)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
		}
		fmt.Printf("Checked out branch: %s\n", branch)
	}

	// Process repository files
	return r.processDirectory(tempDir, repoURL, branch)
}

// processDirectory walks through a directory and processes all code files
func (r *RepoIndexer) processDirectory(dir, repoURL, branch string) error {
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
		if isBinaryFile(path) {
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

		if containsBinaryData(content) {
			fmt.Printf("Skipping file with binary data: %s\n", relPath)
			skippedCount++
			return nil
		}

		// Process file content
		fmt.Printf("Processing file: %s\n", relPath)
		if err := r.processFile(string(content), path, repoURL, branch); err != nil {
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

// processFile splits a file into chunks and indexes them
func (r *RepoIndexer) processFile(content, filePath, repoURL, branch string) error {
	// Check for binary content first
	if containsBinaryData([]byte(content)) {
		return fmt.Errorf("cannot process binary file")
	}

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
	language := getLanguageFromExtension(filepath.Ext(filePath))

	// Split content into chunks of approximately 1000 tokens
	chunks := splitIntoChunks(content, 1000)
	fmt.Printf("Split into %d chunks\n", len(chunks))

	// Process each chunk
	for i, chunk := range chunks {
		// Get embedding for the chunk
		embedding, err := r.getEmbeddings(chunk)
		if err != nil {
			return fmt.Errorf("failed to get embedding: %w", err)
		}

		// Create code chunk
		codeChunk := CodeChunk{
			Content:    chunk,
			FilePath:   relPath,
			Repository: repository,
			Branch:     branch,
			Language:   language,
			Embedding:  embedding,
		}

		fmt.Printf("Storing chunk for %s, repository %s\n", relPath, repository)

		// Store in vector database
		if err := r.vectorStore.Store(codeChunk); err != nil {
			return fmt.Errorf("failed to store chunk: %w", err)
		}

		if i == 0 || i%10 == 0 {
			fmt.Printf("Indexed chunk %d for file: %s\n", i, relPath)
		}
	}

	return nil
}

// getEmbeddings generates embeddings for the given text using OpenAI's API
func (r *RepoIndexer) getEmbeddings(text string) ([]float32, error) {
	resp, err := r.openAIClient.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{text},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %v", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return resp.Data[0].Embedding, nil
}

// Utility functions

func splitIntoChunks(content string, chunkSize int) []string {
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

	return chunks
}

func getLanguageFromExtension(ext string) string {
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

func isBinaryFile(filename string) bool {
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

func containsBinaryData(content []byte) bool {
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

// Execute processes the repository indexing request
func (r *RepoIndexer) Execute(params map[string]interface{}) (interface{}, error) {
	repoURL, ok := params["repoUrl"].(string)
	if !ok || repoURL == "" {
		return nil, fmt.Errorf("repository URL is required")
	}

	branch, _ := params["branch"].(string)
	if branch == "" {
		branch = "main" // Default to main branch
	}

	err := r.IndexRepository(repoURL, branch)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "success",
		"message": "Repository indexed successfully",
	}, nil
}
