package cmd

import (
	"fmt"
	"os"

	"github.com/ory/dockertest/v4"

	"github.com/tommi2day/gomodules/common"
)

const kmsPort = 18080

var kmsHost = common.GetEnv("KMS_HOST", "127.0.0.1")
var kmsAddress = fmt.Sprintf("http://%s:%d", kmsHost, kmsPort)

// prepareKmsContainer creates a moto mock server Docker Container providing an AWS KMS
// compatible API for testing.
func prepareKmsContainer() (dockertest.ClosableResource, error) {
	if os.Getenv("SKIP_KMS") != "" {
		return nil, fmt.Errorf("skipping KMS Container in CI environment")
	}
	return runMotoContainer("pwcli-kms", "KMS_CONTAINER_NAME", kmsHost, kmsPort, "KMS")
}
