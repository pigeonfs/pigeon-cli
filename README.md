# Pigeon CLI

Official CLI for the [Pigeon](https://github.com/pigeonfs/pigeon) email API. Command shape follows [resend-cli](https://github.com/resend/resend-cli): `pigeon emails send`, `pigeon doctor`.

Built for humans, AI agents, and CI.

## Install

```bash
go install github.com/pigeonfs/pigeon-cli/cmd/pigeon@latest
```

Or from a clone:

```bash
git clone https://github.com/pigeonfs/pigeon-cli.git
cd pigeon-cli
go install ./cmd/pigeon
```

## Quickstart

```bash
export PIGEON_API_KEY=pg_xxxx
# export PIGEON_BASE_URL=https://pigeon.bitscorp.co

pigeon emails send \
  --from "Ada <ada@yourdomain.com>" \
  --to person@example.com \
  --subject "Hello from Pigeon CLI" \
  --text "Sent from my terminal."

pigeon domains list
pigeon doctor
```

Auth (first match wins):

| Priority | Source | How to set |
| --- | --- | --- |
| 1 | `--api-key` | `pigeon --api-key pg_xxx emails send ...` is not used; pass `--api-key` after `send` |
| 2 | `PIGEON_API_KEY` | `export PIGEON_API_KEY=pg_xxx` |
| 3 | Config file | `~/.config/pigeon/credentials.json` `{"api_key":"pg_xxx"}` |

`--api-key` is a flag on `emails send`.

## License

MIT
