# pulse

your ai forgot everything. now it doesn't.

most ai lives in the cloud and forgets you exist the moment you close the tab. pulse runs on your machine. it connects to your accounts. it remembers. next time you ask, it knows what you meant last time.

## what it is

a personal ai agent. local first. sqlite memory. any openai-compatible llm. no cloud, no saas, no account.

- **memory** - sqlite on disk. survives restarts. remembers your focus, your projects, your context. tell it once, it knows forever. or until you forget.
- **connectors** - github is wired. gmail and calendar are next. each connector feeds the same pipeline.
- **llm** - bring your own key. any openai-compatible provider. without one, pattern matching. with one, it reasons.
- **local** - your data never leaves. no telemetry. no phone home. no account.

## architecture

two languages. one binary.

- **rust core** - filtering, memory, llm orchestration, digest. serde, rusqlite, reqwest. the brain.
- **go shell** - http server, web ui, cli, connectors. stdlib only. the hands.

go calls rust. they share sqlite and json. no grpc, no protobuf, no microservices. two languages because each does what it's best at.

## install

```bash
cd rust-core && cargo build --release
cp target/release/pulse-core ~/.pulse/

cd .. && go build -o pulse ./cmd/pulse/
```

## use

```bash
pulse connect github ghp_xxxx
pulse ask "what did i miss"
pulse remember focus "ship v1"
pulse ask "what should i work on"
pulse serve                    # web ui on localhost:9090
```

## why

every notification tool shows you more. inbox, timeline, unread count. none of them know what you care about. they just show more. pulse shows you less. the right less.

and it's yours. not a saas. not a freemium tier. not a data pipeline to someone's dashboard. your agent, your data, your machine.

## license

MIT
