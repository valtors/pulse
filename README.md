# pulse

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)

connect everything. your ai does the rest.

## what

pulse is a personal ai agent that connects to your services and handles things for you. not a framework. not a runtime. a product.

you connect github. you connect gmail. you connect calendar. you ask "what did i miss." pulse reads everything and tells you. you say "remember this." pulse remembers. you say "do this." pulse does it.

your data stays on your machine. single binary. no server. no account. no cloud.

## install

```bash
go install github.com/valtors/pulse/cmd/pulse@latest
```

## use

```bash
pulse                    # start on port 9090
pulse -port 8080         # custom port
pulse -data ~/.pulse      # custom data dir
```

open http://localhost:9090. connect your services. ask.

## api

```bash
# connect a service
curl -X POST localhost:9090/connect \
  -d '{"service":"github","token":"ghp_xxx"}'

# ask
curl -X POST localhost:9090/ask \
  -d '{"input":"what did i miss"}'

# check memory
curl localhost:9090/memory

# health
curl localhost:9090/health
```

## services

- github: notifications, pull requests, issues, discussions
- gmail: unread messages, snippets
- calendar: today's events

more coming.

## how it works

```
┌──────────────────────────────────┐
│  pulse                            │
│                                   │
│  ┌──────────┐  ┌──────────────┐  │
│  │ connect   │  │ memory       │  │
│  │ (github,  │  │ (sqlite,     │  │
│  │  gmail,   │  │  persistent) │  │
│  │  calendar)│  │              │  │
│  └──────────┘  └──────────────┘  │
│                                   │
│  ┌──────────┐  ┌──────────────┐  │
│  │ agent     │  │ http api     │  │
│  │ (do, ask, │  │ (rest, ui)   │  │
│  │  remember)│  │              │  │
│  └──────────┘  └──────────────┘  │
└──────────────────────────────────┘
```

## tech

go. single binary. sqlite (pure-go). no dependencies. your data on your machine.

## license

mit
