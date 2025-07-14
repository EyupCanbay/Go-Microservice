package main

import (
	"context"
	"fmt"
	"microservice/pkg/config"
	_ "microservice/pkg/log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {

	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("app staring ...")

	// server timeout config
	app := fiber.New(fiber.Config{
		IdleTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Concurrency:  256 * 1024,
	})

	app.Get("/healthcheck", func(c fiber.Ctx) error {
		// TODO: check some dependencies
		return c.SendString("OK")
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/", func(c fiber.Ctx) error {
		zap.L().Info("Hello word")
		return c.SendString("hello")
	})

	// start server in a goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			zap.L().Error("failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("server started on port", zap.String("port", appConfig.Port))

	gracefullShutDown(app)
}

func gracefullShutDown(app *fiber.App) {
	// create channel for shutdown signals
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	// wait for shutdown signal
	<-signChan
	zap.L().Info("Shutting down server...")

	// Shutdown with 5 second
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("error during server shutdown", zap.Error(err))
	}

	zap.L().Info("server sracefully stopped")
}

func httpC() {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.google.com", nil)
	if err != nil {
		zap.L().Error("failed to get google", zap.Error(err))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		zap.L().Error("failed to get google", zap.Error(err))
	}
	zap.L().Info("google response", zap.Int("status", resp.StatusCode))

}
