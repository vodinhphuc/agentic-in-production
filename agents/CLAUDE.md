# agents/ — platform configuration, NOT code

This folder holds **platform-side configuration** for each agent experiment: prompts, tool definitions, exported workflow files. There is no Python/Node code here, no service deps.

When adding an agent experiment:
- Create `agents/<name>/` with: `README.md`, `prompts/`, `tools.yaml`, and any platform-export files (e.g. a GoClaw workflow file).
- The corresponding **adapter code** lives in `backend/internal/adapters/<platform>/`, not here.
- See [../docs/adr/0009-phase1-platform-goclaw.md](../docs/adr/0009-phase1-platform-goclaw.md).
