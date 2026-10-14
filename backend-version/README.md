# Book Tracker — Go backend version

This folder is the fuller Book Tracker edition: a public landing page, account creation, login, and a personal library backed by Go. It keeps the original tracker features—add books, search by title or author, change status, delete books, and show reading statistics—and adds server-side validation and user-scoped storage.

## What this version adds

- A landing page with Login and Sign up actions in the top-right navigation
- `login.html` and `signup.html` with client-side feedback backed by Go validation
- Passwords stored as salted, repeated SHA-256 hashes rather than plain text
- HTTP-only, SameSite session cookies
- Per-user book ownership so one account cannot read or modify another account's books
- A `cmd/book-tracker` entry point and `internal/` packages for auth, books, and HTTP handlers
- A separate `users.json` file so account values are visible for this learning project

## Run it

From this folder, run:

```bash
go run ./cmd/book-tracker
```

Then open <http://localhost:8080>.

The server uses only Go's standard library. It serves the landing page at `/`, the auth pages at `/login.html` and `/signup.html`, and the signed-in library at `/library.html`.

It provides these endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/auth/signup` | Create an account and start a session |
| `POST` | `/api/auth/login` | Validate credentials and start a session |
| `POST` | `/api/auth/logout` | Revoke the current session |
| `GET` | `/api/auth/me` | Return the signed-in user |
| `GET` | `/api/books` | List the current user's books, newest first |
| `POST` | `/api/books` | Add a book for the current user |
| `PATCH` | `/api/books/{id}` | Change an owned book's status |
| `DELETE` | `/api/books/{id}` | Delete an owned book |

All book endpoints require a valid session cookie.

## Go architecture

```text
backend-version/
├── cmd/book-tracker/main.go       # Configuration and application wiring
├── internal/auth/auth.go          # Users, validation, password hashing, sessions
├── internal/books/books.go        # Owned-book rules and JSON persistence
├── internal/web/server.go         # HTTP routes, cookies, handlers, responses
├── index.html                     # Public landing page
├── login.html                     # Login form
├── signup.html                    # Signup form
├── library.html                   # Authenticated tracker UI
├── app.js                         # Authenticated library interactions
├── auth.js                        # Login/signup validation and requests
├── books.json                     # Visible book data
└── users.json                     # Visible account data
```

The `cmd/` package wires dependencies together. Domain behavior stays in `internal/auth` and `internal/books`, while `internal/web` translates HTTP requests into those operations.

## Why the data is stored in JSON

The data is stored in [`books.json`](books.json) and [`users.json`](users.json) only to make it easy to see exactly how the values are being stored. JSON is plain text, so the records can be opened and inspected without a database tool or an extra dependency. This is a learning/demo choice, not a claim that JSON files are the best storage option for a production application.

The Go server reads both files when it starts and rewrites the relevant file after an account or book changes. The files are intentionally kept in this folder so the stored values are visible beside the backend code.

## Stored value shape

```json
{
  "id": "generated-uuid",
  "ownerId": "user-uuid",
  "title": "Moby-Dick",
  "author": "Herman Melville",
  "status": "To Read",
  "createdAt": 1770000000000
}
```

`createdAt` is a Unix timestamp in milliseconds and is used to keep the newest books first.

An account record contains an `id`, normalized `email`, `createdAt`, and a `passwordHash`. The password itself is never written to JSON. Sessions are held in memory, so restarting the server requires users to log in again.

## Validation

The browser gives immediate feedback, but the Go handlers repeat the important rules on the server:

- Email must have a valid address shape and is normalized to lowercase.
- Signup passwords need at least 8 characters, including a letter and a number.
- Signup requires matching password and confirmation fields.
- Book title, author, and status are checked before any JSON write.
- Book mutations always check the authenticated owner ID as well as the book ID.

This is an educational single-process implementation. For production authentication, use Argon2id or bcrypt, HTTPS, persistent/expiring sessions, CSRF protection, rate limiting, atomic multi-process file writes, and a database designed for concurrent access.
