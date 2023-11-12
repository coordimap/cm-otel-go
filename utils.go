package cmotel

import (
	"fmt"
	"os"
)

// GetEnvWithPrefix tries to fetch the environment variable value of <prefix><env>. In case it is not set then returns an empty string
func GetEnvWithPrefix(prefix string, env string) string {
	envKey := fmt.Sprintf("%s%s", prefix, env)

	return os.Getenv(envKey)
}

// GetServiceName returns the service name based on the envi variables being set. It first checks if the SERVICE_NAME_PREFIX env variable is set. If so that it generates the name by using the value as a prefix.
// Otherwise if K8S_CLUSTER_NAME and NAMESPACE_NAME have been set then it uses both of them to generate the service name. It also takes into consideration the CM_PREFIX env variable when looking for the previously
// mentioned environment variables. If none has been set then it returns the provided name.
func GetServiceName(name string) string {
	envPrefix := os.Getenv(EnvCmPrefix)
	servicePrefix := GetEnvWithPrefix(envPrefix, EnvServiceNamePrefix)
	if servicePrefix != "" {
		return fmt.Sprintf("%s%s", servicePrefix, name)
	}

	k8sClusterName := GetEnvWithPrefix(envPrefix, EnvK8SClusterName)
	k8sNamespace := GetEnvWithPrefix(envPrefix, EnvK8SNamespaceName)
	if k8sClusterName != "" && k8sNamespace != "" {
		return fmt.Sprintf("%s%s%s", k8sClusterName, k8sNamespace, name)
	}

	return name
}
