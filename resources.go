package cmotel

import (
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// LoadEnvVarsAsResource loads all the known environment variables as resource attributes
func LoadEnvVarsAsResource(prefix string) *resource.Resource {
	foundAttributes := []attribute.KeyValue{}

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)

		switch pair[0] {
		case fmt.Sprintf("%s%s", prefix, EnvNamespaceNameType):
			foundAttributes = append(foundAttributes, semconv.K8SNamespaceName(pair[1]))
			break

		case fmt.Sprintf("%s%s", prefix, EnvNodeIPType):
			foundAttributes = append(foundAttributes, attribute.KeyValue{
				Key:   EnvNodeIPType,
				Value: attribute.StringValue(pair[1]),
			})
			break

		case fmt.Sprintf("%s%s", prefix, EnvNodeNameType):
			foundAttributes = append(foundAttributes, semconv.K8SNodeName(pair[1]))
			break

		case fmt.Sprintf("%s%s", prefix, EnvPodNameType):
			foundAttributes = append(foundAttributes, semconv.K8SPodName(pair[1]))
			break

		case fmt.Sprintf("%s%s", prefix, EnvServiceAccountType):
			foundAttributes = append(foundAttributes, attribute.KeyValue{
				Key:   EnvServiceAccountType,
				Value: attribute.StringValue(pair[1]),
			})
			break

		case fmt.Sprintf("%s%s", prefix, EnvK8SClusterName):
			foundAttributes = append(foundAttributes, attribute.KeyValue{
				Key:   EnvK8SClusterName,
				Value: attribute.StringValue(pair[1]),
			})
			break
		}
	}

	return resource.NewWithAttributes(semconv.SchemaURL, foundAttributes...)
}
