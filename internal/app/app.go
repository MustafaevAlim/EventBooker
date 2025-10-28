package app

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/repository"
	"EventBooker/internal/service"
)

type App struct {
	Handler  *ginext.Engine
	Port     string
	Services *service.Services
	Storage  *repository.Storage
	TgBot    *service.TelegramBot
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgForTgCh := make(chan service.RetryMessage, 100)

	var wg sync.WaitGroup

	server := http.Server{
		Addr:    a.Port,
		Handler: a.Handler,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Logger.Error().Msgf("Run server error: %v", err)
			cancel()
		}
	}()

	scheduler := service.NewSchedulerService(a.Storage.Booking)
	wg.Add(1)
	go func() {
		defer wg.Done()
		scheduler.Start(ctx, 30*time.Second, msgForTgCh)
		defer close(msgForTgCh)

	}()

	if a.TgBot != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.TgBot.ListenUpdated(ctx)
		}()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			go a.TgBot.RetryWorker(ctx, msgForTgCh)
		}()
	}

	<-ctx.Done()

	err := server.Shutdown(ctx)
	if err != nil {
		zlog.Logger.Error().Msgf("Shutdown server error: %v", err)
	}

	wg.Wait()

	err = a.Storage.Close()
	if err != nil {
		zlog.Logger.Error().Msgf("Close storage error: %v", err)

	}

	return nil
}
