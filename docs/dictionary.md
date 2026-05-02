# Dictionary

> **Equally for the user and AI coding agents.** Always check this file before
> using project-specific vocabulary. Add an entry the moment a misunderstanding
> surfaces — that is when it is most valuable.

## Entry format
```markdown
### <Term>
**Aliases:** <other names>
**Definition:** <1–3 sentences>
**Don't confuse with:** <similar terms with different meaning>
**Code:** <pointer>
```
---

### Agent
**Aliases:** none — qualify when you use it
**Definition:** Triple-meaning. Always qualify in writing:
1. **AI coding agent** — Claude Code, Cursor, etc., editing this codebase.
2. **Agent Platform** — external system (Mock for MVP, GoClaw for Phase 1) that runs the LLM tool-using loop.
3. **Agent service / loop** — informal; usually means whichever platform is currently active.
**Don't confuse with:** any of the above with each other.

### Agent Platform
**Aliases:** "the platform"
**Definition:** External system that owns the LLM agent loop, tool execution, conversation history, and platform-side authentication. Mock for Phase 0; GoClaw for Phase 1; later: Dify, NemoClaw, OpenClaw, custom.
**Don't confuse with:** AI coding agent.
**Code:** `backend/internal/adapters/<platform>/`

### Adapter
**Aliases:** `AgentPlatformAdapter`, "platform adapter"
**Definition:** Go implementation that translates between our **canonical event envelope** and a specific Agent Platform's native API. One folder per adapter. All adapters must pass the conformance test suite.
**Don't confuse with:** Trino's "connector" (Trino's vocabulary).
**Code:** `backend/internal/adapters/adapter.go`

### Canonical event envelope
**Aliases:** "canonical events", "the event protocol"
**Definition:** Frozen JSON-Schema-defined event shape used between Backend↔Frontend and Adapter↔Backend. v1 events: `run_started`, `text_delta`, `tool_call_start`, `tool_call_end`, `state_update`, `error`, `run_finished`. Adapters translate platform-native streams INTO this.
**Don't confuse with:** the platform's native event format.
**Code:** `protocols/agent-events.schema.json`

### Session
**Aliases:** none
**Definition:** **Our** concept — Postgres row owned by the backend that binds a logged-in user to a specific agent and to a `platform_conversation_id`.
**Don't confuse with:** Conversation (the platform's concept).
**Code:** `backend/internal/sessions/`

### Conversation
**Aliases:** `platform_conversation_id`
**Definition:** **The platform's** concept — owned and managed externally, holds message history, mapped 1:1 to one of our sessions.
**Don't confuse with:** Session (ours).

### Run
**Aliases:** none
**Definition:** One execution of the agent against one user message. Bracketed by `run_started`/`run_finished`. Has a `run_id` that serves as correlation ID frontend→backend→adapter→audit.
**Don't confuse with:** Session (which contains many runs).
**Code:** `protocols/agent-events.schema.json`

### Tool call
**Aliases:** none
**Definition:** One invocation of a registered tool by the agent during a Run. Has a `call_id` linking `tool_call_start` to `tool_call_end`.
**Don't confuse with:** the function definition (a "tool"); a "tool call" is an invocation.
**Code:** `protocols/agent-events.schema.json`

### Experiment
**Aliases:** none
**Definition:** Self-contained slice that demonstrates one agent concept (Trino tool-use, RAG, …). Lives under `agents/<name>/` (config) with adapter code under `backend/internal/adapters/<platform>/`.
**Don't confuse with:** Phase (a roadmap milestone; Phase 1 *delivers* the Trino experiment).

### Phase
**Aliases:** none
**Definition:** Numbered milestone in the project roadmap. Phase 0 = MVP skeleton; Phase 1 = Trino tool-use POC; Phase 2 = RAG; Phase 3 = multi-agent; Phase 4 = memory; Phase 5 = evals.
**Don't confuse with:** Experiment.

### Detection Finding
**Aliases:** "finding", "alert"
**Definition:** OCSF class 2004. Pre-existing alert/finding row in the toy dataset that the agent can pivot from. Not the agent's output.
**Don't confuse with:** the agent's investigation result.
**Code:** `infra/seed-data/` (Phase 1)

### OCSF class
**Aliases:** "OCSF event class"
**Definition:** Specific event type in the Open Cybersecurity Schema Framework, identified by `class_uid` (e.g. 3002 = Authentication).
**Don't confuse with:** an OOP class. Always say "OCSF class" when ambiguity is possible.
