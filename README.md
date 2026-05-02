# agentic-in-production

Three-tier learning platform for AI agent concepts on cybersecurity (OCSF) data.

See [docs/README.md](docs/README.md) for the learning index, or jump to the
foundational design spec: [docs/superpowers/specs/2026-05-01-agentic-platform-design.md](docs/superpowers/specs/2026-05-01-agentic-platform-design.md).

## Quickstart

First-time setup (toolchains): see [docs/dev-environment.md](docs/dev-environment.md).

    cp .env.example .env
    make install
    make hooks
    make up
    # in another terminal:
    make dev-backend
    # in yet another:
    make dev-frontend
    # browse to http://localhost:5173

## The merge gate

    make verify   # lint + tests + codegen sync check; must pass before commit

## Repo orientation

- [CLAUDE.md](CLAUDE.md) — guidance for AI coding agents
- [docs/dictionary.md](docs/dictionary.md) — shared vocabulary
- [docs/dev-environment.md](docs/dev-environment.md) — installed tools, versions, rebuild steps
- [protocols/](protocols/) — wire contracts (frozen, versioned)
- [docs/adr/](docs/adr/) — architecture decisions
