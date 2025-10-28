package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/api"
	"EventBooker/internal/api/handlers"
	"EventBooker/internal/app"
	"EventBooker/internal/config"
	"EventBooker/internal/repository"
	"EventBooker/internal/service"
)

func main() {
	zlog.Init()

	c, err := config.NewConfig()
	if err != nil {
		zlog.Logger.Fatal().Msg(err.Error())
	}

	pgDSN := fmt.Sprintf("host=%s user=%s password=%s database=%s sslmode=disable",
		c.Postgre.Host, c.Postgre.User, c.Postgre.Password, c.Postgre.DBName)

	pg, err := dbpg.New(pgDSN, nil, &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
	})

	tgbot := service.NewTelegramBot(c.TgBot.Token)
	if tgbot == nil {
		zlog.Logger.Warn().Msg("App starting without telegram bot")
	}

	storage := repository.NewStorage(pg)
	services := service.NewServices(storage)
	handlers := handlers.NewHandlers(services)
	engine := ginext.New("debug")
	api.SetupRoutes(handlers, engine)

	app := app.App{
		Handler:  engine,
		Port:     ":" + c.Server.Port,
		Services: services,
		Storage:  storage,
		TgBot:    tgbot,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	if err := app.Run(ctx); err != nil {
		zlog.Logger.Fatal().Msg(err.Error())
	}

}
