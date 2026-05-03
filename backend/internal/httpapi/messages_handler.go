package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"github.com/phucvd2512/agentic-in-production/backend/internal/agentregistry"
	"github.com/phucvd2512/agentic-in-production/backend/internal/audit"
	"github.com/phucvd2512/agentic-in-production/backend/internal/protocol"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

type MessagesHandler struct {
	Sessions *sessions.Store
	Registry *agentregistry.Store
	Audit    *audit.Log
}

func newRunID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "run_" + hex.EncodeToString(b)
}

func (h MessagesHandler) Send(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	if body.Text == "" {
		http.Error(w, "empty text", 400)
		return
	}

	ctx := r.Context()

	// 1. Look up session + agent.
	sess, err := h.Sessions.Get(ctx, sessionID)
	if err != nil {
		http.Error(w, "session not found", 404)
		return
	}
	agent, err := h.Registry.GetByName(ctx, sess.AgentName)
	if err != nil {
		http.Error(w, "agent not found", 404)
		return
	}

	// 2. Persist user message.
	if _, err := h.Sessions.AppendMessage(ctx, sessionID, "user", body.Text); err != nil {
		http.Error(w, "persist failed", 500)
		return
	}

	// 3. Build adapter on demand (Phase 0: cheap; Phase 1+ may want pooling).
	ad, err := adapters.Build(agent.AdapterKind, agent.Config)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 4. Ensure platform conversation id (StartConversation if absent).
	convID := ""
	if sess.PlatformConvID != nil {
		convID = *sess.PlatformConvID
	}
	if convID == "" {
		convID, err = ad.StartConversation(ctx, adapters.Session{ID: sess.ID, UserID: sess.UserID, AgentName: sess.AgentName})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = h.Sessions.SetPlatformConvID(ctx, sess.ID, convID)
	}

	// 5. Open SSE stream to client.
	w.Header().Set("content-type", "text/event-stream")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", 500)
		return
	}

	runID := newRunID()
	req := adapters.RunRequest{
		RunID:          runID,
		Session:        adapters.Session{ID: sess.ID, UserID: sess.UserID, AgentName: sess.AgentName},
		PlatformConvID: convID,
		UserMessage:    body.Text,
	}
	ch, err := ad.Run(ctx, req)
	if err != nil {
		writeEvent(w, errEvent("internal_error", err.Error()))
		writeEvent(w, finishEvent("error"))
		flusher.Flush()
		return
	}

	// 6. Forward + audit.
	var assistantText string
	for ev := range ch {
		// Validate before forwarding (per ADR-0005 / spec §4.4).
		if vErr := protocol.ValidateEvent(ev.Payload); vErr != nil {
			slog.Warn("adapter emitted invalid event", "err", vErr, "raw", string(ev.Payload))
			writeEvent(w, errEvent("internal_error", "adapter emitted invalid event: "+vErr.Error()))
			writeEvent(w, finishEvent("error"))
			flusher.Flush()
			return
		}
		writeEvent(w, ev.Payload)
		flusher.Flush()
		_ = h.Audit.Append(ctx, sess.ID, runID, ev.Type, ev.Payload)

		if ev.Type == "text_delta" {
			var td struct {
				Text string `json:"text"`
			}
			_ = json.Unmarshal(ev.Payload, &td)
			assistantText += td.Text
		}
	}

	if assistantText != "" {
		_, _ = h.Sessions.AppendMessage(ctx, sessionID, "assistant", assistantText)
	}
}

func writeEvent(w io.Writer, payload []byte) {
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
}

func errEvent(code, msg string) []byte {
	b, _ := json.Marshal(map[string]any{"type": "error", "code": code, "message": msg})
	return b
}

func finishEvent(reason string) []byte {
	b, _ := json.Marshal(map[string]any{"type": "run_finished", "reason": reason})
	return b
}
