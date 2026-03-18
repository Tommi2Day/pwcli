package cmd

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tommi2day/pwcli/test"

	"github.com/tommi2day/gomodules/common"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const vaultRepo = "docker.io/hashicorp/vault"
const vaultRepoTag = "1.21.4"
const postgresRepo = "docker.io/postgres"
const postgresRepoTag = "18.3-bookworm"
const containerTimeout = 120
const rootToken = "pwcli-test"

var vaultcontainerName string
var pgcontainerName string

// preparePostgresContainer create a PostgreSQL Docker Container
func preparePostgresContainer() (container *dockertest.Resource, err error) {
	pgcontainerName = os.Getenv("PG_CONTAINER_NAME")
	if pgcontainerName == "" {
		pgcontainerName = "pwcli-postgres"
	}
	var pool *dockertest.Pool
	pool, err = common.GetDockerPool()
	if err != nil {
		err = fmt.Errorf("cannot attach to docker: %v", err)
		return
	}

	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + postgresRepo

	fmt.Printf("Try to start docker container for %s:%s\n", repoString, postgresRepoTag)
	container, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: repoString,
		Tag:        postgresRepoTag,
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
		},
		Hostname: pgcontainerName,
		Name:     pgcontainerName,
		Mounts: []string{
			test.TestDir + "/docker/postgresql/init:/docker-entrypoint-initdb.d",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		err = fmt.Errorf("error starting postgres docker container: %v", err)
		return
	}

	pool.MaxWait = containerTimeout * time.Second
	// we just wait for the container to be ready, but actually the entrypoint script will run for a few seconds
	// postgres is ready when it starts listening.
	if err = pool.Retry(func() error {
		var db *sql.DB
		pghost, pgport := common.GetContainerHostAndPort(container, "5432/tcp")
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
func getPgHostFromContainer(pgContainer *dockertest.Resource) string {
	if pgContainer != nil && pgContainer.Container != nil && pgContainer.Container.NetworkSettings != nil {
		if bridge, ok := pgContainer.Container.NetworkSettings.Networks["bridge"]; ok {
			return bridge.IPAddress
		}
	}
	return "postgresql"
}

// prepareVaultContainer create a Vault Docker Container
func prepareVaultContainer() (container *dockertest.Resource, pgContainer *dockertest.Resource, err error) {
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
	var pool *dockertest.Pool
	pool, err = common.GetDockerPool()
	if err != nil {
		err = fmt.Errorf("cannot attach to docker: %v", err)
		return
	}

	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + vaultRepo

	fmt.Printf("Try to start docker container for %s:%s\n", repoString, vaultRepoTag)

	// we need to know the internal IP of postgres container for vault to connect to it
	pgHost := getPgHostFromContainer(pgContainer)

	container, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: repoString,
		Tag:        vaultRepoTag,
		Env: []string{
			"VAULT_DEV_ROOT_TOKEN_ID=" + rootToken,
			"VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200",
			"PGHOST=" + pgHost,
			"PGPORT=5432",
		},
		Hostname: vaultcontainerName,
		Name:     vaultcontainerName,
		CapAdd:   []string{"IPC_LOCK"},
		Cmd:      []string{},
		Links:    []string{pgcontainerName + ":postgresql"},
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

	// wait 5s to init container, also needed for startup postgresql
	time.Sleep(20 * time.Second)
	elapsed := time.Since(start)
	fmt.Printf("vault Container is available after %s\n", elapsed.Round(time.Millisecond))

	// provision
	cmdout := ""
	cmd := []string{"/bin/sh", "/vault_provision/vault_init.sh"}
	// PASS PGHOST to the script if needed, though it's already in ENV
	cmdout, _, err = common.ExecDockerCmd(container, cmd)
	if err != nil {
		fmt.Printf("Exec Error %s", err)
	} else {
		fmt.Printf("Cmd:%v\n %s", cmd, cmdout)
	}
	return
}
