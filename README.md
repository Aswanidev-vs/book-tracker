# Book Tracker

A simple, client-side web application for tracking your reading list. Built with plain HTML, CSS, and vanilla JavaScript — no frameworks, no build step, and no backend.

> **Looking for more?** Check out [`backend-version/`](backend-version/) — a Go-powered edition of the same app with a few more features (user accounts, server-side validation, per-user shelves, and secure sessions) and a slightly improved overall experience. See its [README](backend-version/README.md) for details.

## Features

- Landing page with guest access — try the tracker without creating an account via "Continue as guest"
- Add books with title, author, and reading status (To Read, Reading, Finished)
- See your books newest-first
- Search books by title or author (case-insensitive, partial match)
- Change a book's reading status
- Delete books
- Data persists in your browser using `localStorage`
- Responsive layout works on desktop and mobile
- Input validation with helpful error messages
- Accessible design (ARIA labels, keyboard-navigable, focus visible)

## Technologies Used

- HTML5 (semantic elements)
- CSS3 (custom properties, flexbox, grid, responsive design)
- Vanilla JavaScript (ES2022: `crypto.randomUUID()`, `element.toggleAttribute`)
- Browser APIs: `localStorage`, DOM

No external libraries, frameworks, or build tools are required.

## Project File Structure

```
book-tracker/
├── index.html      # Page structure and semantics
├── style.css       # All visual styling (responsive, accessible)
├── app.js          # Application logic: state, rendering, persistence
├── backend-version/ # Go backend version with landing page, auth, and JSON persistence
├── logs.md         # Development log documenting the AI-assisted workflow
└── README.md       # This file
```

## Development Process

This project was built following a structured 16‑prompt specification (see `PROMPTS.md`). Each prompt guided an incremental step—from planning and scaffolding to UI design, feature implementation, persistence, search, UX polish, and testing. The developer log in `logs.md` records what was implemented per prompt and any issues found and fixed.

## Running the Application Locally

1. Clone or download this repository to your local machine.
2. Open `index.html` in your preferred web browser (Chrome, Firefox, Safari, Edge).
   - No web server is required; the file:// protocol works fine.
   - Alternatively, if you prefer to serve it via HTTP, run:
     ```bash
     # If you have Python installed:
     python -m http.server 8000
     # Then visit http://localhost:8000 in your browser
     ```
3. The application is ready to use. Add your first book using the form at the top.

## How Data Is Stored

- Book data lives in a JavaScript array (`books`) that is the single source of truth.
- Whenever a book is added, updated, or deleted, the entire array is saved to the browser's `localStorage` under the key `book-tracker:books`.
- When the application starts, it attempts to load and parse that JSON value:
  - If the key does not exist (first visit), it starts with an empty list.
  - If the value exists but is invalid JSON, a warning is shown and the list starts empty.
  - If the value is valid JSON but contains malformed book entries, only the well-formed books are kept.
- Each book object contains:
  ```js
  {
    id: "cryptographically-unique-UUID",
    title: "string",
    author: "string",
    status: "To Read" | "Reading" | "Finished",
    createdAt: Unix-millisecond-timestamp (for newest-first sort)
  }
  ```
- Because the data lives solely in `localStorage`, it is private to your browser and profile. Clearing site data or using private/incognito mode will not persist the library across sessions.

## Testing Instructions

### Manual Testing
1. **Adding a book**
   - Fill in title and author, select a status, click Add Book.
   - Verify the book appears at the top of the list (newest-first).
   - Verify the form clears and focus returns to the title field.
2. **Validation**
   - Submit with empty title → see "Please enter a title." under the title field.
   - Submit with whitespace-only author → see "Please enter an author." under the author field.
   - Submit with both valid → book adds successfully.
3. **Status change**
   - Click the status dropdown on any book entry and pick a different option.
   - Verify only that book's status chip updates and the underlying data changes.
4. **Delete**
   - Click the Delete button on any book entry.
   - Verify the confirmation dialog appears.
   - Confirm deletion → the book is removed and the list renumbers.
   - Cancel deletion → no change occurs.
   - Delete the last book → the designed empty state appears.
5. **Search**
   - Type into the search input → list filters to show only books whose title or author contains the query (case-insensitive, partial).
   - Clear the input (via the X button or backspace) → full list returns.
   - Verify that adding, deleting, or changing status while a search is active continues to work correctly.
6. **Persistence**
   - Add or modify some books, then refresh the page.
   - Verify your changes are still present.
   - Close and reopen the browser → data remains.
7. **Responsive layout**
   - Resize the browser window narrow → form stacks vertically, search spans full width, list remains readable.
   - Verify no horizontal scrollbar appears at any width.
8. **Keyboard accessibility**
   - Tab through the controls in order: Title → Author → Status → Add Book → Search → Clear (when visible).
   - Press Enter on the focused Add button to submit the form.
   - Arrow keys work inside the status and search dropdowns.
   - Visible focus rings appear on interactive elements.

### Automated Testing (via Playwright)
The application has been validated with the following end-to‑end checks using Playwright (Chromium):
- Add valid book, empty/whitespace validation, very long titles/authors wrap
- Change status independently for each book
- Delete one book and delete the last book
- Search with uppercase/lowercase letters, partial matches, clearing a search
- Refreshing the page after changing data
- Invalid or corrupted `localStorage` data shows a warning without crashing
- Mobile layout (390px width) and keyboard accessibility (tab order, focus visible)
- Edge cases: XSS payloads rendered as inert text, identical titles distinguished by UUID, deletion/status change works while a search is active

## Go Backend Version

The [`backend-version`](backend-version) folder is the fuller edition of this project. It adds a landing page, Login and Sign up links in the top-right navigation, `login.html`, `signup.html`, authenticated sessions, server-side validation, per-user book ownership, and a Go `cmd/` + `internal/` architecture.

It stores books in `backend-version/books.json` and users in `backend-version/users.json` so the stored values are visible and easy to inspect as plain text. JSON is intentionally used for demonstration and learning, not as a production database. Run it from that folder with `go run ./cmd/book-tracker`, then visit `http://localhost:8080`. See [`backend-version/README.md`](backend-version/README.md) for the architecture, routes, validation rules, and JSON shapes.

The original [`index.html`](index.html) also links to this more feature-rich backend edition.

## Known Limitations

- Data is scoped to the browser and profile; there is no synchronization across devices or users.
- Storage is limited by the browser's `localStorage` quota (typically 5–10 MB).
- The application does not include a backend, user accounts, or sync capability.
- No automated test suite (Jest, Vitest, etc.) is included; correctness has been verified via manual and Playwright exploratory testing.
- The interface is intentionally simple; advanced features like editing tags, ratings, reviews, or cover images are not present.

## Future Improvements

- **Optional:** Add a dark‑mode toggle that respects `prefers-color-scheme`.
- **Optional:** Allow users to export/import their library as a JSON file.
- **Optional:** Add a "date finished" field for finished books.
- **Optional:** Include subtle animations (e.g., fade‑in for new books, slide‑out for deleted).
- **Optional:** Provide a print‑friendly stylesheet for generating a physical reading list.
- **Optional:** Integrate with the Open Library or Google Books API to fetch cover images and author bios (would require a proxy or CORS‑enabled endpoint).

---

Built as a learning exercise in vanilla web development. Enjoy tracking your reading journey!
