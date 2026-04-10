package internal

import (
	"context"
	"log/slog"

	"device-service/config"
)

type App struct {
	name string
	cfg  config.Config
}

func New() *App {
	app := &App{
		name: config.AppName,
		cfg:  config.LoadDefault(),
	}
	_ = app.init()
	return app
}

func (a *App) Run(_ context.Context) {
	slog.Info("service skeleton started",
		"service", a.name,
		"role", a.cfg.Role,
		"http_port", a.cfg.HTTP.Port,
		"grpc_port", a.cfg.GRPC.Port,
	)
}

