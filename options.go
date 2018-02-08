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
	"errors"

	"github.com/spf13/cobra"
)

// Most of the following is taken from:
// https://github.com/istio/istio/blob/master/pkg/tracing/options.go

// Options defines the set of options supported by Istio's component tracing package.
type Options struct {
	// URL of zipkin collector (example: 'http://zipkin:9411/api/v1/spans'). This enables tracing for Mixer itself.
	ZipkinURL string

	// URL of jaeger HTTP collector (example: 'http://jaeger:14268/api/traces?format=jaeger.thrift'). This enables tracing for Mixer itself.
	JaegerURL string

	// Whether or not to emit trace spans as log records.
	LogTraceSpans bool
}

// Validate returns whether the options have been configured correctly or an error
func (o *Options) Validate() error {
	// due to a race condition in the OT libraries somewhere, we can't have both tracing outputs active at once
	if o.JaegerURL != "" && o.ZipkinURL != "" {
		return errors.New("can't have Jaeger and Zipkin outputs active simultaneously")
	}

	return nil
}

// TracingEnabled returns whether the given options enable tracing to take place.
func (o *Options) TracingEnabled() bool {
	return o.JaegerURL != "" || o.ZipkinURL != "" || o.LogTraceSpans
}

// AttachCobraFlags attaches a set of Cobra flags to the given Cobra command.
//
// This command attaches the necessary set of flags to expose a CLI to let the
// user control all tracing options.
func AttachCobraFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("trace_zipkin_url", "", "",
		"URL of Zipkin collector (example: 'http://zipkin:9411/api/v1/spans').")

	cmd.PersistentFlags().StringP("trace_jaeger_url", "", "",
		"URL of Jaeger HTTP collector (example: 'http://jaeger:14268/api/traces?format=jaeger.thrift').")

	cmd.PersistentFlags().BoolP("trace_log_spans", "", false,
		"Whether or not to log trace spans.")
}
