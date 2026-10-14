## Development Log 
**Prompt 1 – Plan the application**  
- Explained the three‑file vanilla architecture (index.html, style.css, app.js), single `books` array as source of truth, and localStorage key `book-tracker:books`.

**Prompt 2 – Create project files**  
- Created index.html with heading and links, style.css with a basic reset, and app.js with a console.log.

**Prompt 3 – Design the user interface**  
- Implemented full semantic layout (header, add‑book form, search input, list) and an editorial “Reading Journal” aesthetic (Fraunces/Newsreader/IBM Plex Mono, warm paper background, vermilion accent, hairline rules, hard offset shadows, faint CSS‑only grain). Added a tiny data‑URI favicon to avoid a 404.

**Prompt 4 – Add Book**  
- Added `books` array, `createBook`, form submit handler with per‑field validation (title/author required, trimmed), `showError`/`clearError`, minimal render, and `crypto.randomUUID()` for IDs.

**Prompt 5 – Display the book list**  
- Replaced minimal renderer with reusable `render()` that builds each book entry (number badge, title/author, status chip, Delete button) and the designed empty state with quote and hint.

**Prompt 6 – Change reading status**  
- Added `updateBookStatus` (validates allowed statuses), delegated `change` listener on `.book__status`, and wired it to call `updateBookStatus` then `render` and `saveBooks`.

**Prompt 7 – Delete a book**  
- Added `deleteBook` (splice by ID), delegated `click` listener on `.book__delete`, added `confirm()` guard, and wired it to call `deleteBook` then `render` and `saveBooks`.

**Prompt 8 – Save books with localStorage**  
- Added `storageAvailable()`, `loadBooks()` (defensive JSON parse, validity filter, newest‑first sort), `saveBooks()`, `showStorageWarning()` and `<p id="storage‑warning">` in footer; wired `saveBooks()` into add/status/delete mutations and `loadBooks()` at startup.

**Prompt 9 – Implement live search**  
- Added `searchQuery` state and `searchInput` listener, `getVisibleBooks()` filter, `createNoResultsState()`, wired search to re‑render on input, and added clear‑search button.

**Prompt 10 – Improve the user experience**  
- Split global form‑error into per‑field `#title-error` and `#author-error` with `aria-describedby`; added stats bar (total + per‑status counts + “showing X of Y while searching”), clear‑search button, and ensured long titles/authors wrap via `overflow‑wrap: break-word`.

**Prompt 11 – Structured test review**  
- No code changes; created a test checklist covering all core flows (add, validation, status change, delete, search, persistence, reload, corrupted JSON, mobile layout, keyboard accessibility).

**Prompt 12 – Edge‑case review**  
- Inspected handling of spaces‑only inputs, very long strings, XSS‑like input, invalid status values in storage, corrupted localStorage JSON, identical titles with different IDs, delete/search/status change while a search is active. All behaved correctly; no fixes needed.

**Prompt 13 – Reproduce and fix a bug**  
- A real bug was reported after persistence was introduced: after adding a book, changing its status, and refreshing, the status went back to its old value.  
- Repro: add a book → change status → refresh → status is restored to the pre-change value.  
- Root cause: the status mutation updated the in-memory book object and re-rendered, but the updated collection was not being persisted before refresh.  
- Fix: the reusable `saveBooks()` call is wired into `updateBookStatus()` so every status change is saved immediately after the object changes.  
- Manual verification: change a status to “Reading” or “Finished”, refresh the page, and confirm the status remains.

**Prompt 14 – Final regression test**  
- Verified that all features from Prompts 1‑10 remain usable after the above checks.

**Prompt 15 – Write the README**  
- Created README.md with overview, features, technologies, file structure, run instructions, localStorage explanation, testing instructions, known limitations, and future improvements.

## Issues Found and Fixed

- **Favicon 404 during UI styling (Prompt 3):** Added a tiny inline data‑URI favicon to eliminate the console 404 without affecting visual design.  
- **Stale element references in Playwright testing:** Re‑snapshot the page before querying new refs in test scripts (test‑infra fix, not an app bug).  
- **Persistent stale search query between test runs:** Reset `searchQuery = ""` and clear the search input before each isolated test.  
- **Global form‑error field lacked per‑field association (Prompt 10):** Split into `#title-error` and `#author-error` and linked via `aria-describedby` on the respective inputs.  
- **Initial rendering function lacked status chip styling and delete button (Prompt 5):** Enhanced render to include the colored status chip (via `data‑status`) and a Delete button with confirmation prompt.
