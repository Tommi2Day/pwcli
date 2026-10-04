package cmd

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/ory/dockertest/v4"
	"github.com/tommi2day/gomodules/common"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
)

const motoImage = "docker.io/motoserver/moto"
const motoImageTag = "5.2.3"
const motoContainerTimeout = 120

// runMotoContainer starts a moto mock server Docker Container providing AWS-compatible
// APIs (used for both KMS and Secrets Manager) and waits until it accepts connections.
// containerNameDefault is used unless containerNameEnv is set in the environment.
func runMotoContainer(containerNameDefault, containerNameEnv, host string, port int, label string) (
	resource dockertest.ClosableResource, err error) {
	containerName := os.Getenv(containerNameEnv)
	if containerName == "" {
		containerName = containerNameDefault
	}
	pool, err := common.GetDockerPool()
	if err != nil || pool == nil {
		return nil, fmt.Errorf("cannot attach to docker: %v", err)
	}

	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + motoImage

	ctx := context.Background()
	fmt.Printf("Try to start docker container %s for %s:%s\n", containerName, motoImage, motoImageTag)
	resource, err = pool.Run(ctx, repoString,
		dockertest.WithTag(motoImageTag),
		dockertest.WithHostname(containerName),
		dockertest.WithName(containerName),
		dockertest.WithContainerConfig(func(config *container.Config) {
			if config.ExposedPorts == nil {
				config.ExposedPorts = network.PortSet{}
			}
			config.ExposedPorts[network.MustParsePort("5000/tcp")] = struct{}{}
		}),
		dockertest.WithHostConfig(func(config *container.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyDisabled}
			config.PortBindings = network.PortMap{
				network.MustParsePort("5000/tcp"): {
					{HostIP: netip.MustParseAddr("0.0.0.0"), HostPort: fmt.Sprintf("%d", port)},
				},
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("error starting moto docker container: %v", err)
	}

	address := fmt.Sprintf("http://%s:%d", host, port)
	fmt.Printf("Wait to successfully connect to %s mock with %s (max %ds)...\n", label, address, motoContainerTimeout)
	start := time.Now()
	var c net.Conn
	if err = pool.Retry(ctx, motoContainerTimeout*time.Second, func() error {
		c, err = net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
		if err != nil {
			fmt.Printf("Err:%s\n", err)
		}
		return err
	}); err != nil {
		fmt.Printf("Could not connect to %s mock Container: %s", label, err)
		return resource, err
	}
	_ = c.Close()

	// wait to let moto init
	time.Sleep(5 * time.Second)
	fmt.Printf("Moto %s Container is available after %s\n", label, time.Since(start).Round(time.Millisecond))
	return resource, nil
}
