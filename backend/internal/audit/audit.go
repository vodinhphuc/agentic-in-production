package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

type Entry struct {
	ID         int64           `json:"id"`
	SessionID  string          `json:"session_id"`
	RunID      string          `json:"run_id"`
	Kind       string          `json:"kind"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

type Log struct{ DB *db.DB }

func NewLog(d *db.DB) *Log { return &Log{DB: d} }

func (l *Log) Append(ctx context.Context, sessionID, runID, kind string, payload json.RawMessage) error {
	_, err := l.DB.Pool.Exec(ctx,
		`INSERT INTO audit_log (session_id, run_id, kind, payload) VALUES ($1, $2, $3, $4)`,
		sessionID, runID, kind, payload)
	return err
}

func (l *Log) List(ctx context.Context, sessionID string) ([]Entry, error) {
	rows, err := l.DB.Pool.Query(ctx,
		`SELECT id, session_id, run_id, kind, occurred_at, payload
		 FROM audit_log WHERE session_id=$1 ORDER BY occurred_at, id`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.SessionID, &e.RunID, &e.Kind, &e.OccurredAt, &e.Payload); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
