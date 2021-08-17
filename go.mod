module github.com/benmoss/knative-tracing

go 1.16

require (
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/google/uuid v1.3.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC2
	go.opentelemetry.io/otel/sdk v1.0.0-RC2
	go.opentelemetry.io/otel/trace v1.0.0-RC2
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	google.golang.org/grpc v1.39.0
)

replace github.com/cloudevents/sdk-go/v2 => ../sdk-go/v2
