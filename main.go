package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	tracer          = otel.Tracer("fiber-server")
	meter           = otel.Meter("fiber-server")
	logger          otellog.Logger
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	userLookups     metric.Int64Counter
)

type shutdownFunc func(context.Context) error

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	ctx := context.Background()

	// Initialize OTEL providers
	shutdown := initOTEL(ctx)
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down OTEL: %v", err)
		}
	}()

	app := fiber.New()
	app.Use(otelfiber.Middleware())

	app.Get("/users/:id", handleGetUser)

	log.Fatal(app.Listen(":8080"))
}

func initOTEL(ctx context.Context) shutdownFunc {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("my-service"),
	)

	// Trace
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Metrics
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// Logs
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)
	logger = lp.Logger("fiber-server")

	// Propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// Initialize metrics
	initMetrics()

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down trace provider: %v", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down metric provider: %v", err)
		}
		return lp.Shutdown(ctx)
	}
}

func initMetrics() {
	var err error
	requestCounter, err = meter.Int64Counter("http.server.requests")
	if err != nil {
		log.Fatal(err)
	}

	requestDuration, err = meter.Float64Histogram("http.server.duration")
	if err != nil {
		log.Fatal(err)
	}

	userLookups, err = meter.Int64Counter("app.user.lookups")
	if err != nil {
		log.Fatal(err)
	}
}

func handleGetUser(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	id := c.Params("id")

	emitLog(ctx, "User request received", otellog.String("user_id", id))

	user := getUser(ctx, id)

	// Record metrics
	duration := time.Since(start).Seconds()
	attrs := []attribute.KeyValue{
		attribute.String("endpoint", "/users/:id"),
		attribute.String("user_id", id),
	}
	requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

	emitLog(ctx, "User request completed",
		otellog.String("user_id", id),
		otellog.Float64("duration_seconds", duration))

	return c.JSON(user)
}

func emitLog(ctx context.Context, message string, attrs ...otellog.KeyValue) {
	// Log to stdout
	log.Printf("[OTEL] %s", message)

	// Log to OpenTelemetry
	var record otellog.Record
	record.SetBody(otellog.StringValue(message))
	record.AddAttributes(attrs...)
	logger.Emit(ctx, record)
}

func getUser(ctx context.Context, id string) User {
	_, span := tracer.Start(ctx, "getUser", oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	emitLog(ctx, "Looking up user", otellog.String("user_id", id))

	// Record custom metric
	userLookups.Add(ctx, 1, metric.WithAttributes(attribute.String("user_id", id)))

	return User{
		ID:    id,
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
}
