# Contributing to Pulse

Thanks for your interest in contributing. Pulse is a personal AI agent for notifications, built with Go and Rust.

## Ways to Contribute

- **Bug fixes** - Check issues labeled `bug`
- **Features** - Check issues labeled `enhancement` or `good first issue`
- **Connectors** - Add new service connectors (slack, linear, discord, etc.)
- **Filtering** - Improve notification filtering and priority scoring
- **Memory** - Enhance the SQLite memory store and context gathering
- **Web UI** - Improve the dashboard and digest views
- **Docs** - Improve README, add examples, write guides
- **Tests** - Add test coverage for Go shell and Rust core

## Setup

```bash
git clone https://github.com/valtors/pulse.git
cd pulse

# Rust core
cd rust-core && cargo build --release && cd ..

# Go shell
go build -o pulse ./cmd/pulse/
```

## Development

```bash
# Run tests
go test ./... -count=1

# Run web server
./pulse serve
```

## AI Agent Contribution Guide

If you use AI tools to contribute, document which tools you used and which parts they generated. Keep human review in the loop.

## License

By contributing, you agree that your contributions will be licensed under the MIT license.
