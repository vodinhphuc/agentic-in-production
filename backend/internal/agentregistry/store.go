package agentregistry

import (
	"context"
	"encoding/json"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

type Agent struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	AdapterKind string          `json:"adapter_kind"`
	Config      json.RawMessage `json:"config"`
}

type Store struct{ DB *db.DB }

func NewStore(d *db.DB) *Store { return &Store{DB: d} }

func (s *Store) ListEnabled(ctx context.Context) ([]Agent, error) {
	rows, err := s.DB.Pool.Query(ctx,
		`SELECT name, version, description, enabled, adapter_kind, config
		 FROM agent_registry WHERE enabled = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Agent{}
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.Name, &a.Version, &a.Description, &a.Enabled, &a.AdapterKind, &a.Config); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetByName(ctx context.Context, name string) (Agent, error) {
	row := s.DB.Pool.QueryRow(ctx,
		`SELECT name, version, description, enabled, adapter_kind, config
		 FROM agent_registry WHERE name = $1`, name)
	var a Agent
	err := row.Scan(&a.Name, &a.Version, &a.Description, &a.Enabled, &a.AdapterKind, &a.Config)
	return a, err
}
