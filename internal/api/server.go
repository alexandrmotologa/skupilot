package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/engine"
	"github.com/alexandrmotologa/skupilot/internal/store"
)

// Server coordinates HTTP API routing and embedded static frontend assets.
type Server struct {
	mux    *http.ServeMux
	env    *ServerEnv
	webFS  fs.FS
}

// NewServer initializes the HTTP handler tree.
func NewServer(db *store.DB, eng *engine.Engine, webFS fs.FS) *Server {
	env := &ServerEnv{
		DB:     db,
		Engine: eng,
		Broker: NewEventBroker(),
	}

	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /api/health", env.HealthHandler)
	mux.HandleFunc("GET /api/stats", env.StatsHandler)

	mux.HandleFunc("GET /api/batches", env.ListBatchesHandler)
	mux.HandleFunc("POST /api/batches/upload", env.UploadBatchHandler)
	mux.HandleFunc("GET /api/batches/{id}", env.GetBatchHandler)
	mux.HandleFunc("GET /api/batches/{id}/events", env.BatchSSEHandler)
	mux.HandleFunc("POST /api/batches/{id}/approve-all", env.BatchApproveAllHandler)

	mux.HandleFunc("GET /api/batches/{id}/products", env.ListProductsHandler)
	mux.HandleFunc("GET /api/products/{id}", env.GetProductHandler)
	mux.HandleFunc("PATCH /api/products/{id}", env.UpdateProductHandler)

	mux.HandleFunc("POST /api/batches/{id}/export/shopify", env.ExportShopifyHandler)
	mux.HandleFunc("GET /api/batches/{id}/export/amazon", env.ExportAmazonHandler)
	mux.HandleFunc("GET /api/batches/{id}/export/google", env.ExportGoogleHandler)

	// Embedded static file serving for React Merchandising Studio
	if webFS != nil {
		fileServer := http.FileServer(http.FS(webFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// If request is for an API endpoint that didn't match, return 404 JSON
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeError(w, http.StatusNotFound, "endpoint not found")
				return
			}

			// SPA fallback: check if file exists, else serve index.html
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			if f, err := webFS.Open(path); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			// Rewrite to index.html for client-side routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}

	return &Server{
		mux:   mux,
		env:   env,
		webFS: webFS,
	}
}

// ServeHTTP implements http.Handler with CORS and logging.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for frontend studio dev mode
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}
