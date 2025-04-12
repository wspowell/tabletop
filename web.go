package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"github.com/wspowell/tabletop/filepath"
)

var (
	tmpl *template.Template
	db   *sql.DB
)

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	tmpl, _ = template.ParseGlob(filepath.Format("./templates/*.html"))
}

func serveHtml(ctx context.Context, config Config) {
	router := mux.NewRouter()

	mux.CORSMethodMiddleware(router)

	// gRouter.Use(enableCORS)

	router.HandleFunc("/", Homepage(config.WebsocketPort))

	serveStylesheet(router, "./styles/default.css")

	serveJavascriptFile(router, "./scripts/websocket.js")

	// Serve favicon.
	servePngFile(router, "./images/favicon/android-chrome-192x192.png")
	servePngFile(router, "./images/favicon/android-chrome-512x512.png")
	servePngFile(router, "./images/favicon/apple-touch-icon.png")
	servePngFile(router, "./images/favicon/favicon-16x16.png")
	servePngFile(router, "./images/favicon/favicon-32x32.png")

	// Server d20 border icon.
	servePngFile(router, "./images/d20-outline-border.png")

	// Server all game images.
	serveDataPngFiles(router)

	address := config.ServerHost + ":" + config.ServerPort

	log.Println("web address", address)

	http.ListenAndServe(address, router)
}

func serveDataPngFiles(router *mux.Router) {
	err := filepath.Walk(filepath.Format("./data"),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".png") {
				servePngFile(router, "./"+path)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

// func enableCORS(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Allow requests from any origin
// 		w.Header().Set("Access-Control-Allow-Origin", "*")

// 		// Allow specified HTTP methods

// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

// 		// Allow specified headers
// 		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

// 		// Continue with the next handler
// 		next.ServeHTTP(w, r)
// 	})
// }

func serveStylesheet(router *mux.Router, path string) {
	router.Handle(filepath.Route(path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, r, filepath.Format(path))
	}))
}

func serveJavascriptFile(router *mux.Router, path string) {
	router.Handle(filepath.Route(path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, r, filepath.Format(path))
	}))
}

func servePngFile(router *mux.Router, path string) {
	router.Handle(filepath.Route(path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, r, filepath.Format(path))
	}))
}

func LoginForm(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "login.html", nil)
}

func Homepage(websocketPort string) func(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		WebsocketPort string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "home.html", PageData{
			WebsocketPort: websocketPort,
		})
	}
}
