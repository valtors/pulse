# pulse

connect everything. your ai does the rest.

pulse sits on your machine, connects to your accounts, reads the noise, and tells you what actually matters. 38 notifications become 3. the rest is ci failures you already know about.

## architecture

two languages. one product.

- **rust core** - the brain. filtering, memory (sqlite), llm orchestration. built with serde, rusqlite, reqwest.
- **go shell** - the hands. http server, web ui, cli, connector management. built with stdlib only.

the go binary calls the rust binary. they share a sqlite database and json config. no protobuf, no grpc, no microservices. just two processes that talk.

## install

```bash
# build the rust core
cd rust-core && cargo build --release
cp target/release/pulse-core ~/.pulse/pulse-core

# build the go shell
cd .. && go build -o pulse ./cmd/pulse/
```

## use

```bash
# connect github
pulse connect github ghp_your_token

# get your digest - filtered, prioritized
pulse digest

# ask anything
pulse ask "what did i miss"
pulse ask "what do you know"

# remember things
pulse remember focus "ship pulse v1"
pulse memory

# start the web ui
pulse serve
# open http://localhost:9090

# configure ai (optional, enables smart summaries)
pulse config llm https://api.openai.com/v1 sk-your-key gpt-4o-mini
```

## how filtering works

every notification gets classified:

- **urgent** - you were mentioned, asked to review, or it's your PR/issue
- **important** - activity on things you care about
- **noise** - ci failures, status checks, automated updates

noise is filtered out. you see what matters. 38 becomes 3.

## what's connected

- **github** - notifications, filtering, prioritization
- **gmail** - stub (needs oauth)
- **calendar** - stub (needs oauth)

more connectors coming. the architecture is simple - implement the trait/interface, register it, done.

## philosophy

- local first. your data stays on your machine.
- single binary. no docker, no kubernetes, no saas.
- two languages because each does what it's best at.
- boring tech. sqlite, stdlib, http.
- no telemetry. no phone home. no account.

## license

MIT
