// Portions Copyright 2018 Aspen Mesh Authors.
// Portions Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"fmt"
	"io"
	"time"

	"github.com/golang/glog"
	ot "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-client-go/transport/zipkin"
	zp "github.com/uber/jaeger-client-go/zipkin"
)

// Sample code for configuring & using tracing package
/*
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
*/

// Most of the following is taken from:
// https://github.com/istio/istio/blob/master/pkg/tracing/config.go

type holder struct {
	closer io.Closer
	tracer ot.Tracer
}

var (
	httpTimeout = 5 * time.Second
	sampler     = jaeger.NewConstSampler(true)
	poolSpans   = jaeger.TracerOptions.PoolSpans(false)
	logger      = spanLogger{}
)

// indirection for testing
type newZipkin func(url string, options ...zipkin.HTTPOption) (*zipkin.HTTPTransport, error)

// Configure initializes Istio's tracing subsystem.
//
// You typically call this once at process startup.
// Once this call returns, the tracing system is ready to accept data.
func Configure(serviceName string, options *Options) (io.Closer, error) {
	return configure(serviceName, options, zipkin.NewHTTPTransport)
}

func configure(serviceName string, options *Options, nz newZipkin) (io.Closer, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	reporters := make([]jaeger.Reporter, 0, 3)

	if options.ZipkinURL != "" {
		trans, err := nz(options.ZipkinURL, zipkin.HTTPLogger(logger), zipkin.HTTPTimeout(httpTimeout))
		if err != nil {
			return nil, fmt.Errorf("could not build zipkin reporter: %v", err)
		}
		reporters = append(reporters, jaeger.NewRemoteReporter(trans))
	}

	if options.JaegerURL != "" {
		reporters = append(reporters, jaeger.NewRemoteReporter(transport.NewHTTPTransport(options.JaegerURL, transport.HTTPTimeout(httpTimeout))))
	}

	if options.LogTraceSpans {
		reporters = append(reporters, logger)
	}

	var rep jaeger.Reporter
	if len(reporters) == 0 {
		// leave the default NoopTracer in place since there's no place for tracing to go...
		return holder{}, nil
	} else if len(reporters) == 1 {
		rep = reporters[0]
	} else {
		rep = jaeger.NewCompositeReporter(reporters...)
	}

	// Setup zipkin style tracing
	zipkinPropagator := zp.NewZipkinB3HTTPHeaderPropagator()
	injector := jaeger.TracerOptions.Injector(ot.HTTPHeaders, zipkinPropagator)
	extractor := jaeger.TracerOptions.Extractor(ot.HTTPHeaders, zipkinPropagator)
	opts := []jaeger.TracerOption{poolSpans, injector, extractor}

	tracer, closer := jaeger.NewTracer(serviceName, sampler, rep, opts...)

	// NOTE: global side effect!
	ot.SetGlobalTracer(tracer)

	return holder{
		closer: closer,
		tracer: tracer,
	}, nil
}

func (h holder) Close() error {
	if ot.GlobalTracer() == h.tracer {
		ot.SetGlobalTracer(ot.NoopTracer{})
	}

	if h.closer != nil {
		h.closer.Close()
	}

	return nil
}

type spanLogger struct{}

// Report implements the Report() method of jaeger.Reporter
func (spanLogger) Report(span *jaeger.Span) {
	glog.Infof("Reporting span operation: %s span: %s",
		span.OperationName(), span.String())
}

// Close implements the Close() method of jaeger.Reporter.
func (spanLogger) Close() {}

// Error implements the Error() method of log.Logger.
func (spanLogger) Error(msg string) {
	glog.Error(msg)
}

// Infof implements the Infof() method of log.Logger.
func (spanLogger) Infof(msg string, args ...interface{}) {
	glog.Infof(msg, args...)
}
