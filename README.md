# URL Shortener Template

A two-service Go monorepo: an **identity service** for user registration and auth, and a **URL shortener service** that creates short links, tracks redirect clicks, and exposes per-link analytics. Both services are backed by separate PostgreSQL databases and coordinated via Docker Compose.

This repo is a [LineSpec](https://github.com/livecodelife/linespec) template — the full implementation is regenerated from a provenance manifest by a coding agent.

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [LineSpec](https://github.com/livecodelife/linespec) (for cloning and testing)

Install LineSpec via Homebrew:

```sh
brew tap livecodelife/linespec && brew install linespec
```

Or with Go:

```sh
go install github.com/livecodelife/linespec/cmd/linespec@v3.10.3
```

---

## Clone and regenerate

LineSpec's `clone` command bootstraps a fresh local copy — it initialises the git repo, installs hooks, pulls the `.linespec.yml`, and imports all provenance records from the published manifest:

```sh
linespec clone https://raw.githubusercontent.com/livecodelife/url-shortener-template/refs/heads/main/linespec.manifest.json
cd url-shortener-template
```

Then open your coding agent and point it at the prompt:

```
@prompt.md
```

The agent will read the provenance records and implement the full service from scratch.

---

## Run the stack

From the repo root, start all four containers (two services + two databases):

```sh
docker compose up --build
```

- Identity service → `http://localhost:8081`
- URL shortener service → `http://localhost:8080`

---

## Manual end-to-end walkthrough

Run these requests in order. Set `TOKEN` and `SLUG` from the responses as you go.

### 1. Register a user

```sh
curl -s -X POST http://localhost:8081/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secret"}'
```

### 2. Log in and capture the token

```sh
TOKEN=$(curl -s -X POST http://localhost:8081/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secret"}' \
  | sed 's/.*"token":"\([^"]*\)".*/\1/')
```

### 3. Create a short link

```sh
curl -s -X POST http://localhost:8080/links \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"slug":"gh","destination":"https://github.com"}'
```

Omit `slug` to have one generated automatically.

### 4. Follow the redirect

```sh
curl -v http://localhost:8080/gh
# → 302 Location: https://github.com
```

No auth required — anyone with the link can follow it.

### 5. Check analytics

```sh
curl -s http://localhost:8080/links/gh/analytics \
  -H "Authorization: Bearer $TOKEN"
```

Returns `total_clicks`, `clicks_last_24h`, and a `recent_clicks` list with timestamps, user-agent, referer, and IP.

### 6. List your links

```sh
curl -s http://localhost:8080/links \
  -H "Authorization: Bearer $TOKEN"
```

### 7. Delete a link

```sh
curl -s -X DELETE http://localhost:8080/links/gh \
  -H "Authorization: Bearer $TOKEN"
# → 204 No Content
```

---

## Run the tests

Tests are written as LineSpec integration specs and run against live Docker containers.

```sh
linespec test linespecs/identity/
linespec test linespecs/url-shortener/
```
