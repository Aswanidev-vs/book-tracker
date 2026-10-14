// Book Tracker — state, rendering, add, status change, delete.

const STATUSES = ["To Read", "Reading", "Finished"];

const books = [];

const form = document.getElementById("add-book-form");
const titleInput = document.getElementById("book-title");
const authorInput = document.getElementById("book-author");
const statusInput = document.getElementById("book-status");
const list = document.getElementById("book-list");
const titleError = document.getElementById("title-error");
const authorError = document.getElementById("author-error");
const storageWarning = document.getElementById("storage-warning");
const searchInput = document.getElementById("search-input");
const searchClear = document.getElementById("search-clear");
const statsBox = document.getElementById("stats");

const STORAGE_KEY = "book-tracker:books";

// Live search state: the books array is never mutated by searching.
let searchQuery = "";

// ---- Storage layer (reusable; safe against missing/corrupt data) ----

function storageAvailable() {
  try {
    const probe = "__book_tracker_probe__";
    localStorage.setItem(probe, "1");
    localStorage.removeItem(probe);
    return true;
  } catch {
    return false; // e.g. private mode, disabled storage, quota
  }
}

function showStorageWarning(message) {
  storageWarning.textContent = message;
  storageWarning.hidden = false;
}

// Normalise one raw record from storage into a valid book, or null.
function isValidBook(value) {
  if (typeof value !== "object" || value === null) return false;
  const { id, title, author, status, createdAt } = value;
  return (
    typeof id === "string" &&
    typeof title === "string" &&
    typeof author === "string" &&
    STATUSES.includes(status) &&
    typeof createdAt === "number"
  );
}

function loadBooks() {
  let raw;
  try {
    raw = localStorage.getItem(STORAGE_KEY);
  } catch {
    showStorageWarning("Storage is unavailable — changes will not be saved.");
    return;
  }
  if (raw === null) return; // first run: nothing saved yet

  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch {
    showStorageWarning("Saved data was corrupted and could not be loaded.");
    return;
  }
  if (!Array.isArray(parsed)) {
    showStorageWarning("Saved data was invalid and could not be loaded.");
    return;
  }

  // Keep only well-formed books and keep newest-first ordering.
  books.push(...parsed.filter(isValidBook).sort((a, b) => b.createdAt - a.createdAt));
}

function saveBooks() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(books));
    storageWarning.hidden = true;
  } catch {
    showStorageWarning("Could not save your changes — storage is full or unavailable.");
  }
}

function createBook(title, author, status) {
  return {
    id: crypto.randomUUID(),
    title,
    author,
    status,
    createdAt: Date.now(),
  };
}

function createEmptyState() {
  const item = document.createElement("li");
  item.className = "empty-state";

  const line = document.createElement("p");
  line.className = "empty-state__line";
  line.textContent = "Your shelf is empty.";

  const quote = document.createElement("p");
  quote.className = "empty-state__quote";
  quote.textContent = "\u201CA room without books is like a body without a soul.\u201D";

  const hint = document.createElement("p");
  hint.className = "empty-state__hint";
  hint.textContent = "Add your first book with the form above.";

  item.append(line, quote, hint);
  return item;
}

// Shown when a search matches nothing.
function createNoResultsState(query) {
  const item = document.createElement("li");
  item.className = "empty-state empty-state--search";

  const line = document.createElement("p");
  line.className = "empty-state__line";
  line.textContent = `No books match \u201C${query}\u201D.`;

  const hint = document.createElement("p");
  hint.className = "empty-state__hint";
  hint.textContent = "Try a different title or author, or clear the search.";

  item.append(line, hint);
  return item;
}

function createBookItem(book, index) {
  const item = document.createElement("li");
  item.className = "book";
  item.dataset.id = book.id;
  item.dataset.status = book.status;

  const no = document.createElement("span");
  no.className = "book__no";
  no.textContent = String(index + 1).padStart(2, "0");

  const body = document.createElement("div");
  body.className = "book__body";

  const title = document.createElement("p");
  title.className = "book__title";
  title.textContent = book.title;

  const author = document.createElement("p");
  author.className = "book__author";
  author.textContent = `by ${book.author}`;

  body.append(title, author);

  const controls = document.createElement("div");
  controls.className = "book__controls";

  const statusSelect = document.createElement("select");
  statusSelect.className = "book__status";
  statusSelect.setAttribute("aria-label", `Change status for ${book.title}`);
  for (const status of STATUSES) {
    const option = document.createElement("option");
    option.value = status;
    option.textContent = status;
    statusSelect.append(option);
  }
  statusSelect.value = book.status;

  const deleteButton = document.createElement("button");
  deleteButton.type = "button";
  deleteButton.className = "book__delete";
  deleteButton.textContent = "Delete";

  controls.append(statusSelect, deleteButton);

  item.append(no, body, controls);
  return item;
}

// Books currently visible: everything when search is empty, otherwise
// title/author partial matches (case-insensitive). Never mutates `books`.
function getVisibleBooks() {
  const query = searchQuery.trim().toLowerCase();
  if (query === "") return books;
  return books.filter(
    (book) =>
      book.title.toLowerCase().includes(query) ||
      book.author.toLowerCase().includes(query)
  );
}

function render() {
  list.replaceChildren();

  if (books.length === 0) {
    list.append(createEmptyState());
    return;
  }

  const visible = getVisibleBooks();
  if (visible.length === 0) {
    list.append(createNoResultsState(searchQuery.trim()));
    return;
  }

  visible.forEach((book, index) => {
    list.append(createBookItem(book, index));
  });

  renderStats();
  searchClear.hidden = searchQuery.trim() === "";
}

function showError(message, input) {
  const errorBox = input === titleInput ? titleError : authorError;
  errorBox.textContent = message;
  errorBox.hidden = false;
  input.setAttribute("aria-invalid", "true");
  input.focus();
}

function clearError() {
  titleError.textContent = "";
  titleError.hidden = true;
  authorError.textContent = "";
  authorError.hidden = true;
  titleInput.removeAttribute("aria-invalid");
  authorInput.removeAttribute("aria-invalid");
}

// Stats bar: total plus per-status counts, and a "showing X of Y" note
// while a search is active.
function renderStats() {
  const visible = getVisibleBooks();
  const count = (status) => books.filter((b) => b.status === status).length;

  statsBox.replaceChildren();

  const total = document.createElement("span");
  total.className = "stats__item";
  const searching = searchQuery.trim() !== "" && visible.length !== books.length;
  total.textContent = searching
    ? `Showing ${visible.length} of ${books.length} books`
    : `${books.length} ${books.length === 1 ? "book" : "books"}`;
  statsBox.append(total);

  if (books.length === 0) return;

  for (const [status, cls] of [
    ["To Read", "toread"],
    ["Reading", "reading"],
    ["Finished", "finished"],
  ]) {
    const item = document.createElement("span");
    item.className = "stats__item";
    const dot = document.createElement("span");
    dot.className = `stats__dot stats__dot--${cls}`;
    item.append(dot, document.createTextNode(`${count(status)} ${status.toLowerCase()}`));
    statsBox.append(item);
  }
}

// Update a book's status by ID, ensuring the new status is valid.
function updateBookStatus(bookId, newStatus) {
  if (!STATUSES.includes(newStatus)) {
    // Invalid status; ignore to prevent broken state.
    return;
  }
  const book = books.find(b => b.id === bookId);
  if (!book) return;
  book.status = newStatus;
  render();
  saveBooks();
}

// Listen for status changes via delegation (more efficient than attaching handlers).
list.addEventListener("change", (event) => {
  if (event.target.classList.contains("book__status")) {
    const item = event.target.closest(".book");
    if (!item) return;
    const bookId = item.dataset.id;
    const newStatus = event.target.value;
    updateBookStatus(bookId, newStatus);
  }
});

// Remove one book by its unique ID, leaving all other books untouched.
function deleteBook(bookId) {
  const index = books.findIndex((book) => book.id === bookId);
  if (index === -1) return; // unknown ID: nothing to delete
  books.splice(index, 1);
  render();
  saveBooks();
}

// Delete via delegation: the button's closest .book entry identifies the book.
list.addEventListener("click", (event) => {
  const button = event.target.closest(".book__delete");
  if (!button) return;

  const item = button.closest(".book");
  const bookId = item.dataset.id;

  const book = books.find((b) => b.id === bookId);
  if (!book) return;

  if (!confirm(`Delete "${book.title}" by ${book.author}?`)) return;
  deleteBook(bookId);
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  const title = titleInput.value.trim();
  const author = authorInput.value.trim();
  if (title === "") {
    showError("Please enter a title.", titleInput);
    return;
  }
  if (author === "") {
    showError("Please enter an author.", authorInput);
    return;
  }
  clearError();
  books.unshift(createBook(title, author, statusInput.value));
  form.reset();
  titleInput.focus();
  render();
  saveBooks();
});

// Live search: re-filter as the user types; the collection itself is untouched.
searchInput.addEventListener("input", () => {
  searchQuery = searchInput.value;
  render();
});

// Clear-search button: empties the field and restores the full list.
searchClear.addEventListener("click", () => {
  searchInput.value = "";
  searchQuery = "";
  render();
});

if (!storageAvailable()) {
  showStorageWarning("Storage is unavailable — changes will not be saved.");
}

loadBooks(); // restore saved books before the first paint of the list
render();
