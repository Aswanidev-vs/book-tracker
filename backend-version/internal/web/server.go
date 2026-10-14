package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"book-tracker-backend/internal/auth"
	"book-tracker-backend/internal/books"
)

const sessionCookieName = "book_tracker_session"

func init() {
	_ = mime.AddExtensionType(".css", "text/css; charset=utf-8")
	_ = mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".html", "text/html; charset=utf-8")
	_ = mime.AddExtensionType(".json", "application/json; charset=utf-8")
	_ = mime.AddExtensionType(".svg", "image/svg+xml")
}

type Server struct {
	auth      *auth.Store
	books     *books.Store
	staticDir string
}

func NewServer(authStore *auth.Store, bookStore *books.Store, staticDir string) *Server {
	return &Server{auth: authStore, books: bookStore, staticDir: staticDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/signup", s.signup)
	mux.HandleFunc("/api/auth/login", s.login)
	mux.HandleFunc("/api/auth/logout", s.logout)
	mux.HandleFunc("/api/auth/me", s.me)
	mux.HandleFunc("/api/books", s.booksCollection)
	mux.HandleFunc("/api/books/", s.bookResource)
	mux.HandleFunc("/library", s.servePage("library.html"))
	mux.HandleFunc("/library/", s.servePage("library.html"))
	mux.HandleFunc("/login", s.servePage("login.html"))
	mux.HandleFunc("/login/", s.servePage("login.html"))
	mux.HandleFunc("/signup", s.servePage("signup.html"))
	mux.HandleFunc("/signup/", s.servePage("signup.html"))
	mux.Handle("/", s.staticFiles())
	return securityHeaders(mux)
}

// securityHeaders wraps the mux with baseline protective headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defense-in-depth against content-type sniffing.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Disallow embedding the app in frames or iframes (clickjacking).
		w.Header().Set("X-Frame-Options", "DENY")
		// Do not leak the referrer URL to other origins.
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Restrict browser features and ensure scripts/styles load from same-origin and Google Fonts.
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; "+
				"font-src 'self' https://fonts.gstatic.com data:; "+
				"img-src 'self' data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var input struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Request body must be valid JSON.")
		return
	}
	if input.Password != input.ConfirmPassword {
		writeError(w, http.StatusBadRequest, "Passwords do not match.")
		return
	}
	user, err := s.auth.Signup(input.Email, input.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	if err := s.startSession(w, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Account created, but the session could not be started.")
		return
	}
	writeJSON(w, http.StatusCreated, toPublicUser(user))
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Request body must be valid JSON.")
		return
	}
	user, err := s.auth.Login(input.Email, input.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	if err := s.startSession(w, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not start the session.")
		return
	}
	writeJSON(w, http.StatusOK, toPublicUser(user))
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		s.auth.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	user, ok := s.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Please sign in to continue.")
		return
	}
	writeJSON(w, http.StatusOK, toPublicUser(user))
}

func (s *Server) booksCollection(w http.ResponseWriter, r *http.Request) {
	user, ok := s.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Please sign in to continue.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.books.List(user.ID))
	case http.MethodPost:
		var input struct {
			Title  string `json:"title"`
			Author string `json:"author"`
			Status string `json:"status"`
		}
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, "Request body must be valid JSON.")
			return
		}
		book, err := s.books.Create(user.ID, input.Title, input.Author, input.Status)
		if err != nil {
			writeBookError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, book)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) bookResource(w http.ResponseWriter, r *http.Request) {
	user, ok := s.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Please sign in to continue.")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/books/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "Book not found.")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var input struct {
			Status string `json:"status"`
		}
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, "Request body must be valid JSON.")
			return
		}
		book, err := s.books.UpdateStatus(user.ID, id, input.Status)
		if err != nil {
			writeBookError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, book)
	case http.MethodDelete:
		if err := s.books.Delete(user.ID, id); err != nil {
			writeBookError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w, http.MethodPatch, http.MethodDelete)
	}
}

func (s *Server) currentUser(r *http.Request) (auth.User, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return auth.User{}, false
	}
	userID, ok := s.auth.UserIDForSession(cookie.Value)
	if !ok {
		return auth.User{}, false
	}
	user, ok := s.auth.FindUser(userID)
	return user, ok
}

func (s *Server) startSession(w http.ResponseWriter, userID string) error {
	token, err := s.auth.CreateSession(userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (s *Server) servePage(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		http.ServeFile(w, r, filepath.Join(s.staticDir, name))
	}
}

func (s *Server) staticFiles() http.Handler {
	fileServer := http.FileServer(http.Dir(s.staticDir))
	allowed := map[string]bool{
		"":             true,
		"index.html":   true,
		"library.html": true,
		"login.html":   true,
		"signup.html":  true,
		"style.css":    true,
		"app.js":       true,
		"auth.js":      true,
		"favicon.ico":  true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		if requestPath == "favicon.ico" {
			favPath := filepath.Join(s.staticDir, "favicon.ico")
			if _, err := os.Stat(favPath); err != nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		if !allowed[requestPath] {
			http.NotFound(w, r)
			return
		}

		// Prevent directory listing when requesting root "/" without index.html
		if requestPath == "" {
			if _, err := os.Stat(filepath.Join(s.staticDir, "index.html")); err != nil {
				http.Error(w, "index.html not found in "+s.staticDir, http.StatusInternalServerError)
				return
			}
		}

		// Explicitly set correct MIME type headers for CSS and JS to avoid Windows registry issues
		if strings.HasSuffix(requestPath, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(requestPath, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		fileServer.ServeHTTP(w, r)
	})
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"createdAt"`
}

func toPublicUser(user auth.User) userResponse {
	return userResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, auth.ErrInvalidCredentials) {
		status = http.StatusUnauthorized
	} else if errors.Is(err, auth.ErrEmailAlreadyExists) {
		status = http.StatusConflict
	}
	writeError(w, status, err.Error())
}

func writeBookError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, books.ErrNotFound) {
		status = http.StatusNotFound
	}
	writeError(w, status, err.Error())
}

func methodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
