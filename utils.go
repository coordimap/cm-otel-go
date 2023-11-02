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
