module github.com/open-telemetry/opentelemetry-collector-contrib/processor/ipresolverprocessor

go 1.17

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.51.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/collector v0.51.0
	go.opentelemetry.io/collector/pdata v0.51.0
	go.opentelemetry.io/collector/semconv v0.51.0
	go.uber.org/zap v1.21.0
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal => ../../internal/coreinternal
