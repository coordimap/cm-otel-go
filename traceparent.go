package cmotel

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// ParseTraceParent parses a traceParent header and converts it to a SpanContextConfig
func ParseTraceParent(traceParent string) (context.Context, error) {
	// The traceparent header has the format: "00-<trace-id>-<span-id>-<trace-flags>"
	parts := strings.Split(traceParent, "-")

	if len(parts) != 4 || parts[0] != "00" {
		return context.TODO(), fmt.Errorf("Invalid traceparent header: %s", traceParent)
	}

	traceID, err := trace.TraceIDFromHex(parts[1])
	if err != nil {
		return context.TODO(), err
	}

	spanID, err := trace.SpanIDFromHex(parts[2])
	if err != nil {
		return context.TODO(), err
	}

	traceFlags := trace.TraceFlags(0)
	if parts[3] != "" {
		flagsInt, err := parseHexByte(parts[3])
		if err != nil {
			return context.TODO(), err
		}
		traceFlags = trace.TraceFlags(flagsInt)
	}

	return trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
		TraceState: trace.TraceState{},
		Remote:     true,
	})), nil
}

func parseHexByte(hexByte string) (byte, error) {
	// Parse a hex byte, e.g., "01" to 1
	if len(hexByte) != 2 {
		return 0, fmt.Errorf("Invalid hex byte: %s", hexByte)
	}
	var result byte
	_, err := fmt.Sscanf(hexByte, "%02x", &result)
	return result, err
}
