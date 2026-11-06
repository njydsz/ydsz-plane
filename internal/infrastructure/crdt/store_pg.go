// Package crdt — PostgreSQL 文档存储。
package crdt

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PGStore PostgreSQL 文档存储。
type PGStore struct {
	db *pgxpool.Pool
}

// NewPGStore 构造存储。
func NewPGStore(db *pgxpool.Pool) *PGStore { return &PGStore{db: db} }

// Load 读取文档。
func (s *PGStore) Load(ctx context.Context, docID string) ([]byte, error) {
	var state []byte
	err := s.db.QueryRow(ctx, `
		SELECT state FROM crdt_documents WHERE doc_id = $1`, docID).Scan(&state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// Save 保存文档（幂等 upsert）。
func (s *PGStore) Save(ctx context.Context, docID string, state []byte) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO crdt_documents (doc_id, state, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (doc_id) DO UPDATE SET
			state = EXCLUDED.state,
			updated_at = now()`, docID, state)
	return err
}
