package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/danzelVash/diplom_mirea_vps/internal/config"
	"github.com/danzelVash/diplom_mirea_vps/internal/httpapi"
	"github.com/danzelVash/diplom_mirea_vps/internal/service"
	"github.com/danzelVash/diplom_mirea_vps/internal/store"
)

type App struct {
	cfg    config.Config
	server *http.Server
}

func New() (*App, error) {
	cfg := config.Load()

	st, err := store.Open(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	svc := service.New(st)
	handler := httpapi.New(svc)

	return &App{
		cfg: cfg,
		server: &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: handler,
		},
	}, nil
}

func (a *App) Run() error {
	log.Printf("core-backend listening on %s", a.cfg.ListenAddr)
	return a.server.ListenAndServe()
}
