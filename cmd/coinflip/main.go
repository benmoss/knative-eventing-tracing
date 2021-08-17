package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var (
	id     = uuid.New().String()
	tracer = otel.Tracer("coinflip")
)

func receive(ctx context.Context, event cloudevents.Event) *cloudevents.Event {
	ctx, span := tracer.Start(ctx, "receive")
	defer span.End()

	// log.Printf("Received CloudEvent:\n%s", event)

	reply := event.Clone()
	reply.SetType("coinflip")
	reply.SetSource("knative.dev/eventing/cmd/coinflip/" + id)
	reply.SetExtension("flip", flip(ctx))
	return &reply
}

func flip(ctx context.Context) string {
	_, span := tracer.Start(ctx, "flip")
	defer span.End()
	coin := []string{
		"heads",
		"tails",
	}

	result := coin[rand.Intn(len(coin))]
	span.AddEvent("flipped", trace.WithAttributes(attribute.String("result", result)))

	return result
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ce, err := cloudevents.NewClientHTTP(cloudevents.WithMiddleware(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "receive")
	}))
	if err != nil {
		log.Fatalf("failed to create CloudEvent client, %s", err)
	}

	shutdown := initProvider()
	defer shutdown()

	log.Fatal(ce.StartReceiver(context.Background(), receive))
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

// copypasta from https://github.com/open-telemetry/opentelemetry-go/blob/6da20a272765e2878d50b752e8f9e53850ea8331/example/otel-collector/main.go
func initProvider() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(os.Getenv("HOSTNAME"))))
	handleErr(err, "failed to create resource")

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(os.Getenv("OTLP_TRACE_ENDPOINT")),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	handleErr(err, "failed to create trace exporter")

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
	}
}
