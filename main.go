package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var logger otellog.Logger
var fooCounter metric.Int64Counter

func emitLog(ctx context.Context, message string) {
	var record otellog.Record
	record.SetBody(otellog.StringValue(message))
	logger.Emit(ctx, record)
}

func main() {
	// init otel logger
	ctx := context.Background()

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(os.Getenv("SERVICE_NAME"))))
	if err != nil {
		log.Fatal(err)
	}

	exporter, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpointURL(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")), otlploggrpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))
	defer loggerProvider.Shutdown(ctx)

	logger = loggerProvider.Logger("main")

	// init otel metrics
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpointURL(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")), otlpmetricgrpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	defer meterProvider.Shutdown(ctx)

	meter := meterProvider.Meter("main")
	fooCounter, _ = meter.Int64Counter("foo_requests")

	// init gin
	r := gin.Default()

	r.GET("/foo", func(c *gin.Context) {
		emitLog(ctx, "foo endpoint called")
		fooCounter.Add(ctx, 1)

		resp, _ := http.Get("http://localhost:8080/bar")
		defer resp.Body.Close()
		c.JSON(200, gin.H{"message": "foo called bar"})
	})

	r.GET("/bar", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "bar"})
	})

	r.Run(":8080")
}
