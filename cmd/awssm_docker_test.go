package cmd

import (
	"fmt"
	"os"

	"github.com/ory/dockertest/v4"
	"github.com/tommi2day/gomodules/common"
)

const motoPort = 18081

var motoHost = common.GetEnv("AWSSM_HOST", "127.0.0.1")
var motoAddress = fmt.Sprintf("http://%s:%d", motoHost, motoPort)

// prepareMotoContainer creates a moto mock server Docker Container providing an AWS
// Secrets Manager compatible API for testing.
func prepareMotoContainer() (dockertest.ClosableResource, error) {
	if os.Getenv("SKIP_AWSSM") != "" {
		return nil, fmt.Errorf("skipping Secrets Manager Container in CI environment")
	}
	return runMotoContainer("pwcli-awssm", "AWSSM_CONTAINER_NAME", motoHost, motoPort, "Secrets Manager")
}
