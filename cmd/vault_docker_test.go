package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tommi2day/pwcli/test"

	"github.com/tommi2day/gomodules/common"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const repo = "docker.io/hashicorp/vault"
const repoTag = "1.18.4"
const containerTimeout = 120
const rootToken = "pwcli-test"

var vaultcontainerName string

// prepareVaultContainer create an Oracle Docker Container
func prepareVaultContainer() (container *dockertest.Resource, err error) {
	if os.Getenv("SKIP_VAULT") != "" {
		err = fmt.Errorf("skipping Vault Container in CI environment")
		return
	}
	vaultcontainerName = os.Getenv("VAULT_CONTAINER_NAME")
	if vaultcontainerName == "" {
		vaultcontainerName = "pwcli-vault"
	}
	var pool *dockertest.Pool
	pool, err = common.GetDockerPool()
	if err != nil {
		err = fmt.Errorf("cannot attach to docker: %v", err)
		return
	}

	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + repo

	fmt.Printf("Try to start docker container for %s:%s\n", repoString, repoTag)
	container, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: repoString,
		Tag:        repoTag,
		Env: []string{
			"VAULT_DEV_ROOT_TOKEN_ID=" + rootToken,
			"VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200",
		},
		Hostname: vaultcontainerName,
		Name:     vaultcontainerName,
		CapAdd:   []string{"IPC_LOCK"},
		Cmd:      []string{},
		// ExposedPorts: []string{"8200"},
		/*
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8200": {
					{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", port)},
				},
			},
		*/
		Mounts: []string{
			test.TestDir + "/docker/vault_provision:/vault_provision",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		err = fmt.Errorf("error starting vault docker container: %v", err)
		return
	}

	pool.MaxWait = containerTimeout * time.Second
	vaulthost, vaultport := common.GetContainerHostAndPort(container, "8200/tcp")
	address := fmt.Sprintf("http://%s:%d", vaulthost, vaultport)
	fmt.Printf("Wait to successfully connect to Vault with %s (max %ds)...\n", address, containerTimeout)
	start := time.Now()
	if err = pool.Retry(func() error {
		var resp *http.Response
		//nolint gosec
		resp, err = http.Get(address)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK:%s", resp.Status)
		}
		return nil
	}); err != nil {
		fmt.Printf("Could not connect to Vault Container: %s", err)
		return
	}

	// wait 5s to init container
	time.Sleep(5 * time.Second)
	elapsed := time.Since(start)
	fmt.Printf("vault Container is available after %s\n", elapsed.Round(time.Millisecond))

	// provision
	cmdout := ""
	cmd := []string{"/vault_provision/vault_init.sh"}
	cmdout, _, err = common.ExecDockerCmd(container, cmd)
	if err != nil {
		fmt.Printf("Exec Error %s", err)
	} else {
		fmt.Printf("Cmd:%v\n %s", cmd, cmdout)
	}
	return
}
