// Package indexer walks a directory of Markdown files and indexes them into DuckDB.
package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"lite-rag/internal/config"
	"lite-rag/internal/database"
	"lite-rag/internal/normalizer"
	"lite-rag/pkg/chunker"
)

// Embedder is the subset of llm.Client required by the Indexer.
// Defined as an interface to allow test doubles.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// Indexer walks a directory tree and indexes Markdown files into DuckDB.
type Indexer struct {
	db             *database.DB
	emb            Embedder
	ch             *chunker.Chunker
	embeddingModel string
}

// New creates an Indexer configured from cfg.
// embeddingModel is stored per document so that queries can be filtered to
// chunks produced by the same model, avoiding cross-model vector comparisons.
func New(db *database.DB, emb Embedder, embeddingModel string, cfg config.RetrievalConfig) *Indexer {
	return &Indexer{
		db:             db,
		emb:            emb,
		ch:             chunker.New(cfg.ChunkSize, cfg.ChunkOverlap),
		embeddingModel: embeddingModel,
	}
}

// IndexDir walks dir recursively and indexes all *.md files.
// Per-file errors are logged and do not abort the overall run.
func (idx *Indexer) IndexDir(ctx context.Context, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		if err := idx.indexFile(ctx, path); err != nil {
			slog.Warn("index file failed", "path", path, "error", err)
		}
		return nil
	})
}

// indexFile indexes a single Markdown file. It is a no-op when the file's
// content has not changed since the last index run (idempotency).
func (idx *Indexer) indexFile(ctx context.Context, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	fileHash := sha256Hex(raw)

	// Skip unchanged files (same content and same embedding model).
	_, existingHash, existingModel, err := idx.db.FindDocumentByPath(path)
	if err != nil {
		return fmt.Errorf("find document: %w", err)
	}
	if existingHash == fileHash && existingModel == idx.embeddingModel {
		slog.Debug("skipping unchanged file", "path", path)
		return nil
	}

	// Normalize, then chunk.
	normalized := normalizer.Normalize(string(raw))
	chunks := idx.ch.Chunk(normalized)

	slog.Debug("chunked file", "path", path, "chunks", len(chunks))

	// Build the texts to embed, applying the nomic task prefix.
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = "search_document: " + normalizer.StripMarkdown(c.Content)
	}

	// Generate embeddings (empty slice is fine — db rows will have nil blobs).
	var embeddings [][]float32
	if len(texts) > 0 {
		embeddings, err = idx.emb.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("embed %s: %w", path, err)
		}
	}

	// Assemble DB rows.
	docID := documentID(path, fileHash)
	doc := database.DocumentRow{
		ID:             docID,
		FilePath:       path,
		FileHash:       fileHash,
		TotalChunks:    len(chunks),
		IndexedAt:      time.Now().UTC(),
		EmbeddingModel: idx.embeddingModel,
	}

	dbChunks := make([]database.ChunkRow, len(chunks))
	for i, c := range chunks {
		var emb []float32
		if i < len(embeddings) {
			emb = embeddings[i]
		}
		dbChunks[i] = database.ChunkRow{
			ID:          chunkID(docID, i),
			DocumentID:  docID,
			ChunkIndex:  i,
			HeadingPath: c.HeadingPath,
			Content:     c.Content,
			Embedding:   emb,
		}
	}

	if err := idx.db.ReplaceDocument(doc, dbChunks); err != nil {
		return fmt.Errorf("replace document %s: %w", path, err)
	}

	slog.Info("indexed", "path", path, "chunks", len(chunks))
	return nil
}

// ── ID helpers ─────────────────────────────────────────────────────────────

func documentID(filePath, fileHash string) string {
	h := sha256.Sum256([]byte(filePath + ":" + fileHash))
	return hex.EncodeToString(h[:])
}

func chunkID(docID string, chunkIndex int) string {
	h := sha256.Sum256([]byte(docID + ":" + strconv.Itoa(chunkIndex)))
	return hex.EncodeToString(h[:])
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
