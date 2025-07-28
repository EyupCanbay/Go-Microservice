package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"microservice/app/healthcheck"
	"microservice/app/product"
	"microservice/infra/couchbase"
	"microservice/pkg/config"
	_ "microservice/pkg/log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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
	return func(c *fiber.Ctx) error {
		var req R

		if err := c.BodyParser(&req); err != nil {
			if !errors.Is(err, fiber.ErrUnprocessableEntity) {
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body: " + err.Error()})
		}

		if err := c.QueryParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query parameters: " + err.Error()})
		}

		if err := c.ParamsParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid URL parameters: " + err.Error()})
		}

		if err := c.ReqHeaderParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid header parameters: " + err.Error()})
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		res, err := handler.Handle(ctx, &req)
		if err != nil {
			zap.L().Error("handler execution failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an internal error occurred"})
		}

		return c.Status(fiber.StatusOK).JSON(res)
	}
}

func main() {

	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("app starting ...")

	tp := initTracer()
	httpClient := httpC()
	couchbaseRpository := couchbase.NewCouchbaseRepository(tp)

	getProducthandler := product.NewGetProductHandler(couchbaseRpository, httpClient)
	createProductHandler := product.NewCreateProductHandler(couchbaseRpository)
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

	app.Use(otelfiber.Middleware())

	// type save structure is provided
	// type save bir yapı oluşturuldu
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/healthcheck", Handle[healthcheck.HealthcheckRequest, healthcheck.HealthcheckResponse](healthCheckHandler))
	app.Get("/products/:id", Handle[product.GetProductRequest, product.GetProductResponse](getProducthandler))
	app.Post("/products", Handle[product.CreateProductRequest, product.CreateProductResponse](createProductHandler))

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

func httpC() *http.Client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(transport),
	}

	return httpClient
}

func initTracer() *sdktrace.TracerProvider {

	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("localhost:4318"),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("microservice-go"),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
