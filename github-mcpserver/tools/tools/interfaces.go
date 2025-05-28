package tools

// VectorStoreInterface defines the interface for vector store operations
type VectorStoreInterface interface {
	Store(chunk CodeChunk) error
	Search(query []float32, repository, branch string, limit int) ([]CodeChunk, error)
}

// VectorSearchToolInterface defines the interface for vector search operations
type VectorSearchToolInterface interface {
	Execute(params map[string]interface{}) (interface{}, error)
}
