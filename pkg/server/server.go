package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/stdlib"
	"golang.org/x/sync/errgroup"

	"github.com/Ser9unin/WeatherForecast/config"
	"github.com/Ser9unin/WeatherForecast/pkg/api"
	"github.com/Ser9unin/WeatherForecast/pkg/db/repository"
	openweather "github.com/Ser9unin/WeatherForecast/pkg/external"
	"go.uber.org/zap"
)

func Run() {
	fmt.Println("LAUNCH")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)
	logger.Info("reading config")

	cfgDB := config.NewDBConnectionCfg()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode= %s",
		cfgDB.HostAddress,
		cfgDB.HostPort,
		cfgDB.User,
		cfgDB.Password,
		cfgDB.DBName,
		cfgDB.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("unable to start db: ", zap.Error(err))
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		logger.Error("context cancelled", zap.Error(err))
	}

	storage := repository.New(db)
	api := api.NewAPI(storage, logger)
	router := api.NewRouter()

	logger.Info("запускается работа с сервисом openweather")
	go func() {
		openweathercfg := config.NewOpenWeatherAPIID()

		// запускаем подключение к внешнему сервису и загрузку прогнозов в БД
		newOpenWeatherConnect := openweather.NewOpenWeatherAPI(storage, logger)
		newOpenWeatherConnect.OpenWeatherRun(ctx, openweathercfg.APIID)

		// параллельное асинхронное обновление данных по прогнозу раз в 15 минут
		newOpenWeatherConnect.ParallelConcurrentUpd(ctx, openweathercfg.APIID)
	}()

	// запускаем сервер
	srvcfg := config.NewServerCfg()
	srv := &http.Server{
		Addr:    srvcfg.Port,
		Handler: router,
	}

	logger.Info("запускается http сервер")

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return srv.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return srv.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
}
