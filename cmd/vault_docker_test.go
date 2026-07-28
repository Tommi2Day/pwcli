package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"time"

	"github.com/tommi2day/gomodules/common"

	"github.com/tommi2day/pwcli/test"

	"github.com/moby/moby/api/types/container"
	"github.com/ory/dockertest/v4"
)

const vaultRepo = "docker.io/hashicorp/vault"
const vaultRepoTag = "1.21.4"
const postgresRepo = "docker.io/library/postgres"
const postgresRepoTag = "18.3-bookworm"
const containerTimeout = 120
const rootToken = "pwcli-test"

var vaultcontainerName string
var pgcontainerName string

// preparePostgresContainer create a PostgreSQL Docker Container
func preparePostgresContainer() (resource dockertest.ClosableResource, err error) {
	pgcontainerName = os.Getenv("PG_CONTAINER_NAME")
	if pgcontainerName == "" {
		pgcontainerName = "pwcli-postgres"
	}
	var pool dockertest.ClosablePool
	pool, err = common.GetDockerPool()
	if err != nil {
		err = fmt.Errorf("cannot attach to docker: %v", err)
		return
	}

	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + postgresRepo

	ctx := context.Background()
	fmt.Printf("Try to start docker container for %s:%s\n", repoString, postgresRepoTag)
	resource, err = pool.Run(ctx, repoString,
		dockertest.WithTag(postgresRepoTag),
		dockertest.WithEnv([]string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
		}),
		dockertest.WithHostname(pgcontainerName),
		dockertest.WithName(pgcontainerName),
		dockertest.WithMounts([]string{
			test.TestDir + "/docker/postgresql/init:/docker-entrypoint-initdb.d",
		}),
		dockertest.WithHostConfig(func(config *container.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyDisabled}
		}),
	)

	if err != nil {
		err = fmt.Errorf("error starting postgres docker container: %v", err)
		return
	}

	// we just wait for the container to be ready, but actually the entrypoint script will run for a few seconds
	// postgres is ready when it starts listening.
	if err = pool.Retry(ctx, containerTimeout*time.Second, func() error {
		var db *sql.DB
		pghost, pgport := common.GetContainerHostAndPort(resource, "5432/tcp")
		connStr := fmt.Sprintf("host=%s port=%d user=postgres password=postgres dbname=postgres sslmode=disable", pghost, pgport)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
		return db.Ping()
	}); err != nil {
		fmt.Printf("Could not connect to Postgres Container: %s", err)
		return
	}

	// wait to let the init scripts run
	time.Sleep(10 * time.Second)
	fmt.Printf("Postgres Container is available\n")
	return
}

// getPgHostFromContainer returns the internal IP of the postgres container, falling back to hostname.
func getPgHostFromContainer(pgContainer dockertest.Resource) string {
	if pgContainer != nil {
		if netSettings := pgContainer.Container().NetworkSettings; netSettings != nil {
			if bridge, ok := netSettings.Networks["bridge"]; ok {
				return bridge.IPAddress.String()
			}
		}
	}
	return "postgresql"
}

// prepareVaultContainer create a Vault Docker Container
func prepareVaultContainer() (resource dockertest.ClosableResource, pgContainer dockertest.ClosableResource, err error) {
	if os.Getenv("SKIP_VAULT") != "" {
		err = fmt.Errorf("skipping Vault Container in CI environment")
		return
	}

	pgContainer, err = preparePostgresContainer()
	if err != nil {
		err = fmt.Errorf("could not start postgres container: %v", err)
		return
	}

	vaultcontainerName = os.Getenv("VAULT_CONTAINER_NAME")
	if vaultcontainerName == "" {
		vaultcontainerName = "pwcli-vault"
	}
	var pool dockertest.ClosablePool
	pool, err = common.GetDockerPool()
	if err != nil {
		err = fmt.Errorf("cannot attach to docker: %v", err)
		return
	}

	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + vaultRepo

	ctx := context.Background()
	fmt.Printf("Try to start docker container for %s:%s\n", repoString, vaultRepoTag)

	// we need to know the internal IP of postgres container for vault to connect to it
	pgHost := getPgHostFromContainer(pgContainer)

	resource, err = pool.Run(ctx, repoString,
		dockertest.WithTag(vaultRepoTag),
		dockertest.WithEnv([]string{
			"VAULT_DEV_ROOT_TOKEN_ID=" + rootToken,
			"VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200",
			"PGHOST=" + pgHost,
			"PGPORT=5432",
		}),
		dockertest.WithHostname(vaultcontainerName),
		dockertest.WithName(vaultcontainerName),
		dockertest.WithCmd([]string{}),
		dockertest.WithMounts([]string{
			test.TestDir + "/docker/vault_provision:/vault_provision",
		}),
		dockertest.WithHostConfig(func(config *container.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyDisabled}
			config.CapAdd = []string{"IPC_LOCK"}
			config.Links = []string{pgcontainerName + ":postgresql"}
		}),
	)

	if err != nil {
		err = fmt.Errorf("error starting vault docker container: %v", err)
		return
	}

	vaulthost, vaultport := common.GetContainerHostAndPort(resource, "8200/tcp")
	address := fmt.Sprintf("http://%s:%d", vaulthost, vaultport)
	fmt.Printf("Wait to successfully connect to Vault with %s (max %ds)...\n", address, containerTimeout)
	start := time.Now()
	if err = pool.Retry(ctx, containerTimeout*time.Second, func() error {
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

	// wait 5s to init container, also needed for startup postgresql
	time.Sleep(20 * time.Second)
	elapsed := time.Since(start)
	fmt.Printf("vault Container is available after %s\n", elapsed.Round(time.Millisecond))

	// provision
	cmdout := ""
	cmd := []string{"/bin/sh", "/vault_provision/vault_init.sh"}
	// PASS PGHOST to the script if needed, though it's already in ENV
	cmdout, _, err = common.ExecDockerCmd(resource, cmd)
	if err != nil {
		fmt.Printf("Exec Error %s", err)
	} else {
		fmt.Printf("Cmd:%v\n %s", cmd, cmdout)
	}
	return
}
