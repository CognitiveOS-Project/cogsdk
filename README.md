# cogsdk — CognitiveOS Universal System SDK

The mandatory universal Go SDK for [CognitiveOS](https://cognitive-os.org).
`cogsdk` is the single source of truth for shared concerns that every
component reimplements today: JSON framing, schema validation, daemon and CPM
IPC, and environment inspection.

See [ADR-011](https://github.com/CognitiveOS-Project/product-specs/blob/main/adr/ADR-011-universal-system-sdk.md) for the decision and rationale.

## Layout

```
client/   Typed clients (implemented)
  mcp/      MCP JSON-RPC 2.0 framing, tool-call wrappers, schema validation
  daemon/   cognitiveosd IPC client, system codes, envelope types
  cpm/      CPM RPC client — only path to nested .cgp tools
  env/      cgroup, systemd, config/env resolution
adk/      Agent toolkit (staged)
  agent/  intent/  memory/  prompt/
cdk/      Builder tooling (staged)
  builder/  cloudinit/  target/
```

## Install

```bash
go get github.com/CognitiveOS-Project/cogsdk
```

## Development

```bash
make build    # go build ./...
make test     # run tests
make lint     # go vet
```

## License

MIT — see [LICENSE](LICENSE).
