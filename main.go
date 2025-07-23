package main

import (
	"context"
	"errors"
	"fmt"
	"microservice/app/healthcheck"
	"microservice/app/product"
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

type Request any
type Response any

type HandlerInterface[R any, Res any] interface {
	Handle(ctx context.Context, req *R) (*Res, error)
}

// context propagation yapıldı
// timeout yapılırken bu contexi geçmemiz gerekirdi
func Handle[R any, Res any](handler HandlerInterface[R, Res]) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req R

		if err := c.Bind().Body(&req); err != nil && errors.Is(err, fiber.ErrUnprocessableEntity) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.Bind().Query(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query parameters: " + err.Error()})
		}

		if err := c.Bind().URI(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid Url parameters: " + err.Error()})
		}

		if err := c.Bind().Header(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid header parameters: " + err.Error()})
		}

		ctx, cancel := context.WithTimeout(c.Context(), time.Second)
		defer cancel()
		
		res, err := handler.Handle(ctx, &req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(res)
	}
}

func main() {

	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("app starting ...")

	producthandler := product.NewProductHandler()
	healthCheckHandler := healthcheck.NewHealthCheckHandler()

	// server timeout config
	//uygulamanın sağlıklı çalışabilmesi için hem clienta hem servere
	//time out eklendi
	app := fiber.New(fiber.Config{
		IdleTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Concurrency:  256 * 1024,
	})

	// type save structure is provided
	// type save bir yapı oluşturuldu
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/healthcheck", Handle[healthcheck.HealthcheckRequest, healthcheck.HealthcheckResponse](healthCheckHandler))
	app.Get("/products", Handle[product.GetProductRequest, product.GetProductResponse](producthandler))

	// start server in a goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			zap.L().Error("failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("server started on port", zap.String("port", appConfig.Port))

	gracefulShutDown(app)
}

// sağlıklı inebilmesi için gracefull shut down eklendi
func gracefulShutDown(app *fiber.App) {
	// create channel for shutdown signals
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	// wait for shutdown signal
	<-signChan
	zap.L().Info("Shutting down server...")

	// Shutdown with 5 second timeout
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("error during server shutdown", zap.Error(err))
	}

	zap.L().Info("server gracefully stopped")
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
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		zap.L().Error("failed to get google", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	zap.L().Info("google response", zap.Int("status", resp.StatusCode))
}
