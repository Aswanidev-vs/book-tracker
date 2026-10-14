# Book Tracker — Backend Edition

A Go-powered edition of the Book Tracker with user accounts, server-side validation, per-user shelves, and secure session handling. The frontend is plain HTML/CSS/vanilla JavaScript; the backend is a small, dependency-free Go HTTP server that persists data to human-readable JSON files.

## Project Structure

```
backend-version/
├── cmd/book-tracker/main.go      # Entry point: flags, store wiring, HTTP server
├── internal/
│   ├── auth/auth.go              # Users, password hashing, sessions
│   ├── books/books.go            # Book collection, validation, persistence
│   └── web/server.go             # HTTP routing, handlers, security headers
├── index.html                    # Landing page
├── login.html / signup.html      # Auth pages
├── library.html                  # Authenticated book shelf
├── style.css                     # Shared styling
├── app.js                        # Library page logic
├── auth.js                       # Login/signup page logic
├── users.json                    # User records (created at runtime)
└── books.json                    # Book records (created at runtime)
```

## Running

```bash
cd backend-version
go run ./cmd/book-tracker
# → http://localhost:8080
```

The server resolves its static files and JSON data files relative to the running executable (falling back to the working directory for `go run`), so it can be launched from any directory without accidentally serving the wrong folder. Flags are available for customization:

- `-addr` — listen address (default `:8080`)
- `-static-dir` — directory containing the frontend files
- `-data-dir` — directory containing `users.json` and `books.json`

## Features

- **Landing page** with links to sign up or log in
- **User accounts** — sign up, log in, log out; each user has a private shelf
- **Add books** with title, author, and reading status (To Read, Reading, Finished)
- **Newest-first** ordering, persisted across server restarts
- **Change status** and **delete** individual books
- **Server-side validation** with clear error messages
- **Session cookies** that survive page reloads (24-hour lifetime)
- **Graceful degradation** — the UI shows a warning if the backend is unreachable

## Implemented Security Measures

### Input Validation

- **Email** — trimmed, lowercased, and checked against a strict pattern; duplicates rejected with `409`.
- **Password** — minimum 8 characters with at least one letter and one number; enforced server-side (client-side mirrors the rules for UX).
- **Books** — title and author are trimmed and rejected if empty; status is whitelisted to exactly `To Read`, `Reading`, or `Finished`.
- **JSON decoding** — `DisallowUnknownFields` rejects unexpected fields; a second decode rejects trailing garbage; bodies are capped at 1 MB via `io.LimitReader`.
- **Content type** — the API only accepts `application/json`; form-encoded or other payloads get `400`.

### Password Storage

- **Salted hashing** — each password gets a 16-byte CSPRNG salt.
- **Iterated SHA-256** — 120,000 rounds (PBKDF2-style construction), stored as `salt$hash` in hex.
- **Constant-time comparison** — `subtle.ConstantTimeCompare` prevents timing attacks on verification.
- **File permissions** — `users.json` is written with mode `0600` (owner-only).

### Sessions & Cookies

- **Opaque tokens** — 64-character CSPRNG hex; only the SHA-256 hash is stored server-side, so a leaked data file cannot be replayed as a session.
- **Cookie flags** — `HttpOnly` (inaccessible to JavaScript), `SameSite=Lax` (blocks cross-site request forgery on state-changing methods), `Path=/`.
- **Server-side expiry** — sessions expire after 24 hours; expired tokens are removed on lookup.
- **Logout** — deletes the token server-side and clears the cookie (`MaxAge=-1`), verified to invalidate the session immediately.

### XSS Protection

- **No XSS sinks** — the frontend renders exclusively with `textContent` and DOM APIs; there is no `innerHTML`, `insertAdjacentHTML`, `eval`, or `document.write` anywhere.
- **JSON encoding** — Go's `encoding/json` escapes `<`, `>`, and `&` by default, so stored payloads like `<script>alert(1)</script>` come back as `\u003cscript\u003e...`.
- **Content type** — all API responses are `application/json`, so reflected input is never interpreted as HTML.