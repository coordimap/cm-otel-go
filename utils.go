package cmotel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// GetEnvWithPrefix tries to fetch the environment variable value of <prefix><env>. In case it is not set then returns an empty string
func GetEnvWithPrefix(prefix string, env string) string {
	envKey := fmt.Sprintf("%s%s", prefix, env)

	return os.Getenv(envKey)
}

// GetServiceName returns the service name based on the unique service prefix.
func GetServiceName(name string) string {
	return fmt.Sprintf("%s.%s", GetUniqueServicePrefix(), name)
}

// GetUniqueServicePrefix returns the prefix that will be used to name the service and any other remote services that are being called. It takes into account the environment variables being set.
// It first checks if the SERVICE_NAME_PREFIX env variable is set. If so that it generates the name by using the value as a prefix. Otherwise if K8S_CLUSTER_NAME and NAMESPACE_NAME have been set
// then it uses both of them to generate the service name. It also takes into consideration the CM_PREFIX env variable when looking for the previously mentioned environment variables. If none has been set then it returns the provided name.
func GetUniqueServicePrefix() string {
	envPrefix := os.Getenv(EnvCmPrefix)
	servicePrefix := GetEnvWithPrefix(envPrefix, EnvServiceNamePrefix)
	if servicePrefix != "" {
		return servicePrefix
	}

	k8sClusterName := GetEnvWithPrefix(envPrefix, EnvK8SClusterName)
	k8sNamespace := GetEnvWithPrefix(envPrefix, EnvK8SNamespaceName)
	if k8sClusterName != "" && k8sNamespace != "" {
		return fmt.Sprintf("%s.%s", k8sClusterName, k8sNamespace)
	}

	return ""
}

// InitDefaultTracerProvider initialize the default provider
func InitDefaultTracerProvider(exporter sdktrace.SpanExporter) error {
	prefix := GetEnvWithPrefix("", EnvCmPrefix)
	serviceName := GetServiceName(GetEnvWithPrefix(prefix, EnvServiceName))

	res, errRes := resource.Merge(
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
		LoadEnvVarsAsResource(prefix),
	)

	if errRes != nil {
		return errRes
	}

	tracer := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithMaxExportBatchSize(50)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracer)

	return nil
}

// LoadRESTEndpointAtributes returns an array of KeyValues related to the given request object
func LoadRESTEndpointAtributes(r *http.Request) []attribute.KeyValue {
	foundAttributes := []attribute.KeyValue{}

	foundAttributes = append(foundAttributes,
		semconv.HTTPMethod(r.Method),
		semconv.HTTPURL(r.RequestURI),
		semconv.HTTPRoute(r.URL.Path),
	)

	var buf bytes.Buffer
	tee := io.TeeReader(r.Body, &buf)

	// load json
	contents, errReadAll := io.ReadAll(tee)
	if errReadAll == nil {
		jsonKeys, errJSONKeys := ExtractJSONKeys(contents)
		if errJSONKeys == nil {
			// TODO: add the array as an attribute
			fmt.Println(jsonKeys)
		}
	}

	return []attribute.KeyValue{}
}

// ExtractJSONKeys returns a slice of all the keys found in a json byte slice. If there are nested objects they are of the format parentkey.childkey
func ExtractJSONKeys(jsonBytes []byte) ([]string, error) {
	var jsonData interface{}

	err := json.Unmarshal(jsonBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	var result []string

	extractKeysRecursive(jsonData, "", &result)

	return result, nil
}

func extractKeysRecursive(data interface{}, prefix string, result *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPrefix := prefix + key + "."
			extractKeysRecursive(value, newPrefix, result)
		}
	case []interface{}:
		for i, value := range v {
			newPrefix := prefix + strconv.Itoa(i) + "."
			extractKeysRecursive(value, newPrefix, result)
		}
	default:
		*result = append(*result, prefix[:len(prefix)-1]) // remove the trailing dot
	}
}
