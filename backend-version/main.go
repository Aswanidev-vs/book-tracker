package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"book-tracker-backend/internal/auth"
	"book-tracker-backend/internal/books"
	"book-tracker-backend/internal/web"
)

func isAppDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "library.html")); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return true
		}
	}
	return false
}

// appDir returns the directory that holds the frontend files and JSON data files.
// It checks the current directory, walks up parent directories, checks subdirectories,
// and checks the executable path to reliably locate the backend-version folder.
func appDir() string {
	// 1. Direct check in current directory
	if isAppDir(".") {
		if abs, err := filepath.Abs("."); err == nil {
			return abs
		}
		return "."
	}

	// 2. Check "backend-version" subdirectory (if executed from repository root)
	if isAppDir("backend-version") {
		if abs, err := filepath.Abs("backend-version"); err == nil {
			return abs
		}
		return "backend-version"
	}

	// 3. Walk up from current working directory
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 6; i++ {
			if isAppDir(dir) {
				return dir
			}
			if isAppDir(filepath.Join(dir, "backend-version")) {
				return filepath.Join(dir, "backend-version")
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// 4. Walk up from executable directory (for compiled binary)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for i := 0; i < 6; i++ {
			if isAppDir(dir) {
				return dir
			}
			if isAppDir(filepath.Join(dir, "backend-version")) {
				return filepath.Join(dir, "backend-version")
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "."
}

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	staticDir := flag.String("static-dir", "", "directory containing the frontend files (defaults to the application directory)")
	dataDir := flag.String("data-dir", "", "directory containing the JSON data files (defaults to the application directory)")
	flag.Parse()

	baseDir := appDir()
	if *staticDir == "" {
		*staticDir = baseDir
	}
	if *dataDir == "" {
		*dataDir = baseDir
	}

	authStore := auth.NewStore(filepath.Join(*dataDir, "users.json"))
	if err := authStore.Load(); err != nil {
		log.Fatalf("load users: %v", err)
	}
	bookStore := books.NewStore(filepath.Join(*dataDir, "books.json"))
	if err := bookStore.Load(); err != nil {
		log.Fatalf("load books: %v", err)
	}

	server := web.NewServer(authStore, bookStore, *staticDir)
	log.Printf("Book Tracker backend running at http://localhost%s", *address)
	log.Printf("Serving static files from: %s", *staticDir)
	log.Printf("JSON data files: %s and %s", filepath.Join(*dataDir, "users.json"), filepath.Join(*dataDir, "books.json"))
	if err := http.ListenAndServe(*address, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
