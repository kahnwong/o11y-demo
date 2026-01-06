package main

import (
	"context"
	"log"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Setup OTEL Logs
	logExporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL("http://localhost:4318/v1/logs"))
	if err != nil {
		log.Fatal(err)
	}
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
	global.SetLoggerProvider(lp)

	// Setup OTEL Traces
	traceExporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpointURL("http://localhost:4318"))
	if err != nil {
		log.Fatal(err)
	}
	tp := trace.NewTracerProvider(trace.WithBatcher(traceExporter))
	otel.SetTracerProvider(tp)

	// Setup OTEL Metrics
	metricExporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpointURL("http://localhost:4318"))
	if err != nil {
		log.Fatal(err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	otel.SetMeterProvider(mp)

	// Create instruments
	meter := otel.Meter("demo-app")
	counter, _ := meter.Int64Counter("requests_total")
	logger := global.GetLoggerProvider().Logger("demo-app")

	app := fiber.New()
	app.Use(otelfiber.Middleware())

	app.Get("/", func(c *fiber.Ctx) error {
		counter.Add(c.Context(), 1)
		var record otellog.Record
		record.SetBody(otellog.StringValue("GET / endpoint accessed"))
		logger.Emit(c.Context(), record)
		log.Println("GET / called")
		return c.JSON(fiber.Map{"message": "Hello World"})
	})

	app.Post("/data", func(c *fiber.Ctx) error {
		counter.Add(c.Context(), 1)
		var record otellog.Record
		record.SetBody(otellog.StringValue("POST /data endpoint accessed"))
		logger.Emit(c.Context(), record)
		log.Println("POST /data called")
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			var errorRecord otellog.Record
			errorRecord.SetBody(otellog.StringValue("Invalid JSON received"))
			logger.Emit(c.Context(), errorRecord)
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		return c.JSON(fiber.Map{"received": body})
	})

	err = app.Listen(":8080")
	if err != nil {
		return
	}
}
