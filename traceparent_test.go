package cmotel

import (
	"reflect"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestParseTraceParent(t *testing.T) {
	type args struct {
		traceParent string
	}
	tests := []struct {
		name    string
		args    args
		want    trace.SpanContext
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				traceParent: "00-12345678901234567890123456789012-3456789012345678-01",
			},
			wantErr: false,
			want: trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    trace.TraceID{},
				SpanID:     trace.SpanID{},
				TraceFlags: trace.TraceFlags(0x01),
				TraceState: trace.TraceState{},
				Remote:     true,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTraceParent(tt.args.traceParent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTraceParent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTraceParent() = %v, want %v", got, tt.want)
			}
		})
	}
}
