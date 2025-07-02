package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

var tpl = template.Must(template.ParseGlob("templates/*.html"))

type App struct{ db *sql.DB }

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or cannot be loaded: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec("SET search_path TO public")
	if err != nil {
		log.Fatalln(err)
	}

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot ping database: %v", err)
	}
	log.Println("Database connection successful")

	// Check current database name
	var dbName string
	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Printf("Cannot get database name: %v", err)
	} else {
		log.Printf("Connected to database: %s", dbName)
	}

	app := &App{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/signup", app.handleSignup)
	mux.HandleFunc("/name", app.requireAuth(app.handleName))

	// handles the static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	srv := &http.Server{
		Addr:              ":8443",
		Handler:           app.secureHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("fullchain.pem", "privkey.pem"))
}

/* ---------- handlers ---------- */

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	// handle POST request for login
	if r.Method == http.MethodPost {
		if !a.validCSRF(r) {
			http.Error(w, "bad csrf", 400)
			return
		}

		var id int
		var hash []byte
		err := a.db.QueryRow(`SELECT id,pw_hash FROM users WHERE username=$1`, r.FormValue("user")).Scan(&id, &hash)
		if err != nil || bcrypt.CompareHashAndPassword(hash, []byte(r.FormValue("pw"))) != nil {
			http.Error(w, "invalid credentials", 401)
			return
		}

		sid := newToken()
		_, _ = a.db.Exec(`INSERT INTO sessions(id,user_id,expires) VALUES($1,$2,now()+interval '12 hours')`, sid, id)

		http.SetCookie(w, &http.Cookie{
			Name:     "sess", Value: sid, Path: "/", HttpOnly: true,
			Secure:   true, SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/name", http.StatusSeeOther)
		return
	}

	// else handle GET request
	var csrf string
	if cookie, err := r.Cookie("csrf"); err == nil {
		// Reuse existing CSRF token if present
		csrf = cookie.Value
	} else {
		// Generate new token only if needed
		csrf = newToken()
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf",
			Value:    csrf,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   true,
		})
	}
	_ = tpl.ExecuteTemplate(w, "index.html", map[string]string{"Csrf": csrf})
}

func (a *App) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		csrf := newToken()
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf",
			Value:    csrf,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   true,
		})
		_ = tpl.ExecuteTemplate(w, "signup.html", map[string]string{"Csrf": csrf})
		return
	}

	if !a.validCSRF(r) {
		http.Error(w, "bad csrf", 400)
		return
	}

	u, pw := r.FormValue("user"), r.FormValue("pw")
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)

	_, err := a.db.Exec(`INSERT INTO users(username,pw_hash) VALUES($1,$2)`, u, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "user exists?", 409)
		} else {
			http.Error(w, "unexpected error: "+err.Error(), 500)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) handleName(w http.ResponseWriter, r *http.Request) {
	a.render("name.html", w)
}

/* ---------- helpers ---------- */

func (a *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid, err := r.Cookie("sess")
		if err != nil {
			http.Redirect(w, r, "/", 302)
			return
		}
		var ok bool
		err = a.db.QueryRow(`SELECT expires>now() FROM sessions WHERE id=$1`, sid.Value).Scan(&ok)
		if err != nil || !ok {
			http.Redirect(w, r, "/", 302)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (a *App) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func (a *App) validCSRF(r *http.Request) bool {
	c, err := r.Cookie("csrf")
	if err != nil {
		log.Printf("CSRF validation failed: cookie not found")
		return false
	}
	formValue := r.FormValue("csrf")
	valid := c.Value == formValue
	if !valid {
		log.Printf("CSRF validation failed: cookie=%s vs form=%s", c.Value, formValue)
	}
	return valid
}

func newToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *App) render(name string, w http.ResponseWriter) {
	_ = tpl.ExecuteTemplate(w, name, nil)
}
