package sessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

type Session struct {
	ID             string
	UserID         string
	AgentName      string
	PlatformConvID *string
	CreatedAt      time.Time
	EndedAt        *time.Time
}

type Message struct {
	ID        string
	SessionID string
	Role      string // "user" | "assistant"
	Text      string
	CreatedAt time.Time
}

type Store struct{ DB *db.DB }

func NewStore(d *db.DB) *Store { return &Store{DB: d} }

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

func (s *Store) Create(ctx context.Context, user, agentName string) (Session, error) {
	id := newID("sess")
	row := s.DB.Pool.QueryRow(ctx,
		`INSERT INTO sessions (id, user_id, agent_name) VALUES ($1, $2, $3)
		 RETURNING id, user_id, agent_name, platform_conversation_id, created_at, ended_at`,
		id, user, agentName)
	var out Session
	err := row.Scan(&out.ID, &out.UserID, &out.AgentName, &out.PlatformConvID, &out.CreatedAt, &out.EndedAt)
	return out, err
}

func (s *Store) Get(ctx context.Context, id string) (Session, error) {
	row := s.DB.Pool.QueryRow(ctx,
		`SELECT id, user_id, agent_name, platform_conversation_id, created_at, ended_at
		 FROM sessions WHERE id = $1`, id)
	var out Session
	err := row.Scan(&out.ID, &out.UserID, &out.AgentName, &out.PlatformConvID, &out.CreatedAt, &out.EndedAt)
	return out, err
}

func (s *Store) SetPlatformConvID(ctx context.Context, id, convID string) error {
	_, err := s.DB.Pool.Exec(ctx,
		`UPDATE sessions SET platform_conversation_id = $1 WHERE id = $2`, convID, id)
	return err
}

func (s *Store) ListMessages(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := s.DB.Pool.Query(ctx,
		`SELECT id, role, text, created_at FROM messages WHERE session_id=$1 ORDER BY created_at`,
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.SessionID = sessionID
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) AppendMessage(ctx context.Context, sessionID, role, text string) (Message, error) {
	id := newID("msg")
	row := s.DB.Pool.QueryRow(ctx,
		`INSERT INTO messages (id, session_id, role, text) VALUES ($1, $2, $3, $4)
		 RETURNING id, session_id, role, text, created_at`, id, sessionID, role, text)
	var m Message
	err := row.Scan(&m.ID, &m.SessionID, &m.Role, &m.Text, &m.CreatedAt)
	return m, err
}
