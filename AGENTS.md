# CognitiveOS Universal SDK (cogsdk)

`cogsdk` is the mandatory universal Go SDK for CognitiveOS. It is the single
source of truth for shared types, JSON framing, schema validation, and typed
clients. See `ADR-011` in the product-specs repo.

## Tier Layout (ADR-011)

| Tier | Package | Status |
|------|---------|--------|
| `client/mcp` | MCP JSON-RPC 2.0 framing, tool-call wrappers, schema validation | implemented |
| `client/daemon` | cognitiveosd IPC client, system codes, envelope types | implemented |
| `client/cpm` | CPM RPC client — only path to nested `.cgp` tools | implemented |
| `client/env` | cgroup, systemd, config/env resolution | implemented |
| `adk/agent` | Agent toolkit | staged |
| `adk/intent` | Intent recognition | staged |
| `adk/memory` | Memory persistence | staged |
| `adk/prompt` | Prompt construction | staged |
| `cdk/builder` | Image/toolchain builder | staged |
| `cdk/cloudinit` | Cloud-init orchestration | staged |
| `cdk/target` | Target provisioning | staged |

## Rules

- **No direct tool execution.** Nested `.cgp` tool invocation is strictly
  mediated by the cpm daemon (ADR-004 validated tool domains). cogsdk ships
  typed clients only.
- **No private variants.** If you need framing, validation, or IPC, use cogsdk —
  do not reimplement it in your repo.
- **SSOT.** Shared JSON Schema definitions live in this module; validate
  in-process (shift-left), not at registry/daemon latency.
- **Keep the public API small.** Package docs are the contract; exported symbols
  need a doc comment.

## Build

```bash
make build    # go build ./...
make test     # run tests
make lint     # go vet
make clean    # remove build artifacts
```

## Development

All repos follow the git workflow defined in root `.opencode/instructions/git-workflow.md`:

- Branch from `development`, not `main`
- Use topic branches: `feature/<name>`, `fix/<name>`, `bugfix/<name>`
- Open a PR to `development` — squash merge after review
- No rebase — prefer `git pull` (merge)
- Commit types: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`

## Dependencies

- `github.com/santhosh-tekuri/jsonschema/v6` — in-process schema validation (`client/mcp`)
- Standard library otherwise

## Cloning Convention

- Use SSH (`git@github.com:`) for development.
- Use HTTPS (`https://github.com/`) for build scripts that clone public dependencies.
