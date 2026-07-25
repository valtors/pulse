# pulse examples

## connect github

```bash
pulse connect github ghp_xxxxxxxxxxxx
```

## get your digest

```bash
pulse digest
```

output:
```
3 notifications worth your attention (out of 41):

[URGENT] @sarah mentioned you in #auth-bug
  "we need your input on the token refresh issue"
  github.com/valtors/relay/issues/42

[REVIEW] PR #38 needs your review
  "fix: handle empty tool list in proxy"
  github.com/valtors/observer/pull/38

[INFO] CI passed on main
  all 4 checks green. you already knew.
  (not included in urgent count)
```

## ask questions

```bash
pulse ask "what did i miss while i was away"
pulse ask "any urgent prs?"
pulse ask "what's the status of the observer issue"
```

## memory

```bash
# tell pulse what you care about
pulse remember focus "ship observer v1 this week"
pulse remember ignore "bot-created prs"
pulse remember project observer "mcp proxy for tool call logging"

# see what it remembers
pulse memory

# forget
pulse forget "old focus"
pulse forget --older-than 30d
```

## web ui

```bash
pulse serve
# open http://localhost:9090
```

## without an llm

pulse works without an api key. it falls back to pattern matching:
- ci failures: matched by repo + branch + status
- mentions: matched by @your-handle
- prs: matched by assignee + review request
- spam: filtered by bot detection

add an llm key for reasoning:
```bash
export OPENAI_API_KEY=sk-xxxx
pulse config llm openai
```
