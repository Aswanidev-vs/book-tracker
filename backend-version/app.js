// Book Tracker backend version — browser UI backed by the Go JSON API.

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
const userEmail = document.getElementById("user-email");
const logoutButton = document.getElementById("logout-button");

let searchQuery = "";

function showServerWarning(message) {
  storageWarning.textContent = message;
  storageWarning.hidden = false;
}

function clearServerWarning() {
  storageWarning.textContent = "";
  storageWarning.hidden = true;
}

async function request(path, options = {}) {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const text = await response.text();
  let payload = null;
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = null;
    }
  }
  if (!response.ok) {
    const error = new Error(payload?.error || `Server error (${response.status})`);
    error.status = response.status;
    throw error;
  }
  return payload;
}

function isValidBook(value) {
  return (
    value &&
    typeof value.id === "string" &&
    typeof value.title === "string" &&
    typeof value.author === "string" &&
    STATUSES.includes(value.status) &&
    typeof value.createdAt === "number"
  );
}

async function loadBooks() {
  try {
    const savedBooks = await request("/api/books");
    books.splice(0, books.length, ...savedBooks.filter(isValidBook));
    clearServerWarning();
    render();
  } catch (error) {
    if (error.status === 401) {
      window.location.replace("login.html");
      return;
    }
    showServerWarning("The Go backend is unavailable — changes cannot be saved.");
  }
}

async function loadSession() {
  try {
    const user = await request("/api/auth/me");
    userEmail.textContent = user.email;
    await loadBooks();
  } catch {
    window.location.replace("login.html");
  }
}

logoutButton.addEventListener("click", async () => {
  logoutButton.disabled = true;
  try {
    await request("/api/auth/logout", { method: "POST" });
  } finally {
    window.location.href = "index.html";
  }
});

function createEmptyState() {
  const item = document.createElement("li");
  item.className = "empty-state";

  const line = document.createElement("p");
  line.className = "empty-state__line";
  line.textContent = "Your shelf is empty.";

  const quote = document.createElement("p");
  quote.className = "empty-state__quote";
  quote.textContent = "“A room without books is like a body without a soul.”";

  const hint = document.createElement("p");
  hint.className = "empty-state__hint";
  hint.textContent = "Add your first book with the form above.";

  item.append(line, quote, hint);
  return item;
}

function createNoResultsState(query) {
  const item = document.createElement("li");
  item.className = "empty-state empty-state--search";

  const line = document.createElement("p");
  line.className = "empty-state__line";
  line.textContent = `No books match “${query}”.`;

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

function getVisibleBooks() {
  const query = searchQuery.trim().toLowerCase();
  if (query === "") return books;
  return books.filter(
    (book) =>
      book.title.toLowerCase().includes(query) ||
      book.author.toLowerCase().includes(query),
  );
}

function render() {
  list.replaceChildren();

  if (books.length === 0) {
    list.append(createEmptyState());
    searchClear.hidden = true;
    renderStats();
    return;
  }

  const visible = getVisibleBooks();
  if (visible.length === 0) {
    list.append(createNoResultsState(searchQuery.trim()));
    renderStats();
    searchClear.hidden = searchQuery.trim() === "";
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

titleInput.addEventListener("input", () => {
  titleError.textContent = "";
  titleError.hidden = true;
  titleInput.removeAttribute("aria-invalid");
});

authorInput.addEventListener("input", () => {
  authorError.textContent = "";
  authorError.hidden = true;
  authorInput.removeAttribute("aria-invalid");
});

function renderStats() {
  const visible = getVisibleBooks();
  const count = (status) => books.filter((book) => book.status === status).length;

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

async function updateBookStatus(bookId, newStatus) {
  if (!STATUSES.includes(newStatus)) return;
  try {
    const updatedBook = await request(`/api/books/${encodeURIComponent(bookId)}`, {
      method: "PATCH",
      body: JSON.stringify({ status: newStatus }),
    });
    const book = books.find((entry) => entry.id === bookId);
    if (book) book.status = updatedBook.status;
    clearServerWarning();
    render();
  } catch (error) {
    if (error.status === 401) {
      window.location.replace("login.html");
      return;
    }
    showServerWarning(error.message);
    render();
  }
}

list.addEventListener("change", (event) => {
  if (!event.target.classList.contains("book__status")) return;
  const item = event.target.closest(".book");
  if (item) updateBookStatus(item.dataset.id, event.target.value);
});

async function deleteBook(bookId) {
  try {
    await request(`/api/books/${encodeURIComponent(bookId)}`, { method: "DELETE" });
    const index = books.findIndex((book) => book.id === bookId);
    if (index !== -1) books.splice(index, 1);
    clearServerWarning();
    render();
  } catch (error) {
    if (error.status === 401) {
      window.location.replace("login.html");
      return;
    }
    showServerWarning(error.message);
  }
}

list.addEventListener("click", (event) => {
  const button = event.target.closest(".book__delete");
  if (!button) return;
  const item = button.closest(".book");
  const book = books.find((entry) => entry.id === item?.dataset.id);
  if (book && confirm(`Delete "${book.title}" by ${book.author}?`)) {
    deleteBook(book.id);
  }
});

form.addEventListener("submit", async (event) => {
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
  try {
    const book = await request("/api/books", {
      method: "POST",
      body: JSON.stringify({ title, author, status: statusInput.value }),
    });
    books.unshift(book);
    form.reset();
    titleInput.focus();
    clearServerWarning();
    render();
  } catch (error) {
    if (error.status === 401) {
      window.location.replace("login.html");
      return;
    }
    showServerWarning(error.message);
  }
});

searchInput.addEventListener("input", () => {
  searchQuery = searchInput.value;
  render();
});

searchClear.addEventListener("click", () => {
  searchInput.value = "";
  searchQuery = "";
  render();
});

loadSession();
