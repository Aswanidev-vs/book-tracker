const form = document.getElementById("auth-form");
const mode = form.dataset.authMode;
const emailInput = document.getElementById("email");
const passwordInput = document.getElementById("password");
const confirmPasswordInput = document.getElementById("confirm-password");
const submitButton = form.querySelector(".auth-submit");
const message = document.getElementById("auth-message");

// Check if user is already logged in; if so, send them straight to library
fetch("/api/auth/me")
  .then(res => res.ok ? res.json() : null)
  .then(user => {
    if (user && user.email) {
      window.location.replace("library.html");
    }
  })
  .catch(() => {});

function setError(input, messageText) {
  const error = document.getElementById(`${input.id}-error`);
  if (!error) return;
  error.textContent = messageText;
  error.hidden = messageText === "";
  input.toggleAttribute("aria-invalid", messageText !== "");
}

function setMessage(messageText) {
  message.textContent = messageText;
  message.hidden = messageText === "";
}

function validEmail(value) {
  return /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(value.trim());
}

function validPassword(value) {
  return value.length >= 8 && /[A-Za-z]/.test(value) && /\d/.test(value);
}

function updatePasswordRequirements() {
  const checks = {
    length: passwordInput.value.length >= 8,
    letter: /[A-Za-z]/.test(passwordInput.value),
    number: /\d/.test(passwordInput.value),
  };
  document.querySelectorAll("[data-requirement]").forEach((item) => {
    const satisfied = checks[item.dataset.requirement];
    item.toggleAttribute("data-satisfied", satisfied);
    item.textContent = `${satisfied ? "✓" : "○"} ${item.textContent.replace(/^[✓○] /, "")}`;
  });
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

// Clear errors dynamically on user typing
emailInput.addEventListener("input", () => setError(emailInput, ""));
passwordInput.addEventListener("input", () => {
  setError(passwordInput, "");
  if (mode === "signup") updatePasswordRequirements();
});
if (confirmPasswordInput) {
  confirmPasswordInput.addEventListener("input", () => setError(confirmPasswordInput, ""));
}

document.querySelectorAll(".password-toggle").forEach((button) => {
  button.addEventListener("click", () => {
    const input = document.getElementById(button.dataset.passwordTarget);
    if (!input) return;
    const showing = input.type === "text";
    input.type = showing ? "password" : "text";
    button.textContent = showing ? "Show" : "Hide";
    button.setAttribute("aria-label", `${showing ? "Show" : "Hide"} password`);
  });
});

if (mode === "signup") {
  updatePasswordRequirements();
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  setMessage("");
  setError(emailInput, "");
  setError(passwordInput, "");
  if (confirmPasswordInput) setError(confirmPasswordInput, "");

  const email = emailInput.value.trim();
  let firstInvalid = null;
  if (!validEmail(email)) {
    setError(emailInput, "Enter a valid email address.");
    firstInvalid = emailInput;
  }
  if (passwordInput.value === "" || (mode === "signup" && !validPassword(passwordInput.value))) {
    setError(passwordInput, mode === "signup" ? "Use at least 8 characters, including a letter and a number." : "Enter your password.");
    firstInvalid ||= passwordInput;
  }
  if (confirmPasswordInput && passwordInput.value !== confirmPasswordInput.value) {
    setError(confirmPasswordInput, "Passwords do not match.");
    firstInvalid ||= confirmPasswordInput;
  }
  if (firstInvalid) {
    firstInvalid.focus();
    return;
  }

  submitButton.disabled = true;
  submitButton.textContent = mode === "signup" ? "Creating shelf…" : "Signing in…";
  try {
    await request(`/api/auth/${mode}`, {
      method: "POST",
      body: JSON.stringify({
        email,
        password: passwordInput.value,
        ...(confirmPasswordInput ? { confirmPassword: confirmPasswordInput.value } : {}),
      }),
    });
    window.location.href = "library.html";
  } catch (error) {
    setMessage(error.message);
    submitButton.disabled = false;
    submitButton.textContent = mode === "signup" ? "Create my shelf" : "Sign in";
  }
});
