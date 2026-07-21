<div align="center">

# pulse

your notifications are ~~fine~~ drowning you.

a personal ai agent that lives on your machine. not in the cloud. not on someone's server. yours.

[landing](https://valtors.github.io/pulse/) - [github](https://github.com/valtors/pulse)

</div>

---

![pulse landing](https://raw.githubusercontent.com/valtors/pulse/gh-pages/assets/landing.png)

## the problem

38 pings a day. 29 are ci failures you already know about. 3 are mentions you'll miss. you don't have a notification problem. you have a filter problem. pulse fixes it.

## what it does

- **remembers you** - sqlite on disk. survives restarts. it knows your focus, your projects, what you care about and what you ignore. tell it once. it knows forever. or until you tell it to forget.
- **connects accounts** - github is wired. gmail and calendar are next. each connector feeds the same pipeline.
- **thinks** - bring your own key. any openai-compatible provider. without an llm, pattern matching. with one, it reasons.
- **stays yours** - your data never leaves. no telemetry. no phone home. no account. sqlite, json, one binary.

## architecture

two languages. one binary.

```
rust core (the brain)          go shell (the hands)
+------------------+          +------------------+
| filtering        |          | http server      |
| sqlite memory    |<--json-->| web ui           |
| llm orchestration|          | cli              |
| digest builder   |          | connectors       |
+------------------+          +------------------+
         |                            |
         +------ shared sqlite -------+
```

go calls rust. they share sqlite and json. no grpc, no protobuf, no microservices. two languages because each does what it's best at. if you can't tell why that's interesting, this isn't for you.

## install

```bash
# rust core
cd rust-core && cargo build --release
cp target/release/pulse-core ~/.pulse/

# go shell
cd .. && go build -o pulse ./cmd/pulse/
```

## cli

```bash
pulse connect github ghp_xxxx     # connect a service
pulse digest                       # get your filtered summary
pulse ask "what did i miss"        # ask anything
pulse remember focus "ship v1"     # store memory
pulse memory                       # show what it remembers
pulse serve                        # web ui on localhost:9090
```

## web ui

![pulse web ui](https://raw.githubusercontent.com/valtors/pulse/gh-pages/assets/ui.png)

```bash
pulse serve
# open http://localhost:9090
```

digest at the top. ask anything. connect services. memory persists across restarts.

## stack

`rust` `go` `sqlite` `serde` `rusqlite` `reqwest` `stdlib` `openai-compatible` `single binary` `no build step`

## why

every notification tool shows you more. inbox, timeline, unread count. none of them know what you care about. they just show more. pulse shows you less. the right less.

and it's yours. not a saas. not a freemium tier. not a data pipeline to someone's dashboard. your agent, your data, your machine.

## license

MIT
