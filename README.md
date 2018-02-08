### Configuring tracing in your application

This package can be used to configure tracing in your application which is
compatible with Istio and Aspen Mesh. It setups up the global opentracing
tracer which can be used by applications to propragate tracing headers or create
new spans.

```go
import (
  "log"

  "github.com/spf13/cobra"
  "github.com/spf13/viper"

  "github.com/aspenmesh/tracing-go"
)

func setupTracing() {
	// Configure Tracing
	tOpts := &tracing.Options{
		ZipkinURL:     viper.GetString("trace_zipkin_url"),
		JaegerURL:     viper.GetString("trace_jaeger_url"),
		LogTraceSpans: viper.GetBool("trace_log_spans"),
	}
	if err := tOpts.Validate(); err != nil {
		log.Fatal("Invalid options for tracing: ", err)
	}
	var tracer io.Closer
	if tOpts.TracingEnabled() {
		tracer, err = tracing.Configure("myapp", tOpts)
		if err != nil {
			tracer.Close()
			log.Fatal("Failed to configure tracing: ", err)
		} else {
			defer tracer.Close()
		}
	}
}
```

Note that most of this package is derived from the Istio
[tracing](https://github.com/istio/istio/tree/master/pkg/tracing) package.
