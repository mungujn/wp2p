package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port              int    `mapstructure:"PORT"  default:"8080"`
	URLPrefix         string `mapstructure:"URL_PREFIX"  default:"/api"`
	Domain            string `mapstructure:"DOMAIN"  default:"github.com/mungujn"`
	CORSAllowedHost   string `mapstructure:"CORS_ALLOWED_HOST"  default:"*"`
}

// New creates a new server
func New(
	cfg Config,
) (*Server, error) {
	// build http server
	httpSrv := http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
	}

	// build Server
	srv := Server{config: cfg}

	srv.setupHTTP(&httpSrv)
	return &srv, nil
}

type Server struct {
	http   *http.Server
	config Config
}

// setupHTTP sets up the http server
func (s *Server) setupHTTP(srv *http.Server) {
	router := s.GetRouter()

	srv.Handler = cors.New(cors.Options{
		AllowedOrigins:     []string{s.config.CORSAllowedHost},
		AllowedMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowedHeaders:     []string{"*"},
		AllowCredentials:   true,
		OptionsPassthrough: false,
	}).Handler(router)

	s.http = srv
}

// GetRouter returns a mux router
func (s *Server) GetRouter() *mux.Router {
	r := mux.NewRouter()

	r.Handle("/files", http.HandlerFunc(s.GetFile)).Methods(http.MethodGet)

	return r
}

// GetFile returns a file
func (s *Server) GetFile(w http.ResponseWriter, r *http.Request) {
	SendEmptyResponse(w, http.StatusOK)
}

// Run runs the web server
func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("http service: begin run")

	go func() {
		defer wg.Done()
		log.Debug("http service: addr=", s.http.Addr)
		err := s.http.ListenAndServe()
		log.Error(err)
		log.Info("http service: end run > ", err)
	}()

	go func() {
		<-ctx.Done()
		sdCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.http.Shutdown(sdCtx)
		if err != nil {
			log.Info("http service shutdown (", err, ")")
		}
	}()
}
