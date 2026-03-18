package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/pwcli/test"
)

func TestVault(t *testing.T) {
	var err error
	var out = ""
	test.InitTestDirs()
	if os.Getenv("SKIP_VAULT") != "" {
		t.Skip("Skip Vault Test in CI")
		return
	}
	vaultContainer, pgContainer, err := prepareVaultContainer()
	defer common.DestroyDockerContainer(vaultContainer)
	defer common.DestroyDockerContainer(pgContainer)
	require.NoErrorf(t, err, "Vault Server not available")
	require.NotNil(t, vaultContainer, "Prepare failed")
	if err != nil || vaultContainer == nil {
		t.Fatal("Vault server not available")
	}
	vaulthost, vaultport := common.GetContainerHostAndPort(vaultContainer, "8200/tcp")
	address := fmt.Sprintf("http://%s:%d", vaulthost, vaultport)
	_ = os.Setenv("VAULT_ADDR", address)
	err = os.Setenv("VAULT_TOKEN", rootToken)
	if err != nil {
		t.Fatalf("cannot set vault environment")
	}
	t.Logf("ADDR=%s, Token=%s", address, rootToken)
	t.Run("CMD vault write", func(t *testing.T) {
		args := []string{
			"vault",
			"write",
			"--logical=false",
			"--info",
			"--unit-test",
			"--mount", "secret",
			"--path", "test",
			"--vault_addr", address,
			"--vault_token", rootToken,
			"{\"password\": \"testpass\"}",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Write command should  not return an error: %s", err)
		assert.Contains(t, out, "Vault Write OK", "Output should not confirm success")
		t.Log(out)
	})

	t.Run("CMD vault read", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--logical=false",
			"--info",
			"--unit-test",
			"--mount", "secret",
			"--path", "test",
			"--vault_addr", address,
			"--vault_token", rootToken,
			"password",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		// Output contains log messages and data. Check for data exactly.
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD vault read json", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--logical=false",
			"--info",
			"--unit-test",
			"--mount", "secret",
			"--path", "test",
			"--json",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		assert.True(t, strings.Contains(out, "{"), "Output should be json")
		t.Log(out)
	})
	viper.Reset()
	jsonOut = false
	t.Run("CMD vault read export", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--logical=false",
			"--info",
			"--unit-test",
			"--mount", "secret",
			"--path", "test",
			"--json=false",
			"--export",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		assert.True(t, strings.Contains(out, "export PASSWORD=\"testpass\""), "Output should be export format")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD vault list", func(t *testing.T) {
		args := []string{
			"vault",
			"list",
			"--info",
			"--unit-test",
			"--mount", "secret",
			"--path", "/",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		expected := "Vault List returned 3 entries"
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Containsf(t, out, expected, "Output should ccontain %s", expected)
	})
	t.Run("CMD vault list demo", func(t *testing.T) {
		args := []string{
			"vault",
			"list",
			"--info",
			"--unit-test",
			"--mount", "secret",
			"--path", "demo",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		expected := "Vault List returned 2 entries"
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Containsf(t, out, expected, "Output should ccontain %s", expected)
	})
	t.Run("CMD vault list empty", func(t *testing.T) {
		args := []string{
			"vault",
			"list",
			"--info",
			"--unit-test",
			"--mount", "",
			"--path", "dummy",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		expected := "Vault List returned 0 entries"
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Containsf(t, out, expected, "Output should ccontain %s", expected)
	})
	viper.Reset()
	t.Run("CMD GetPassword Vault", func(t *testing.T) {
		args := []string{
			"get",
			"--method", "vault",
			"--debug",
			"--unit-test",
			"--config", test.TestData + "/test_pwcli.yaml",
			"--path", "secret/data/test",
			"--entry", "password",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm success")
		t.Log(out)
	})
	t.Run("test removing global options", func(t *testing.T) {
		args := []string{
			"vault",
			"list",
			"--help",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "help command should  not return an error: %s", err)
		assert.Contains(t, out, "--path", "Output should contain path flag")
		assert.NotContains(t, out, "--datadir", "Output should not contain datadir flag")
		t.Log(out)
	})

	t.Run("CMD vault read database role demo-ro", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--info",
			"--unit-test",
			"--logical",
			"--path", "database/creds/demo-ro",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(strings.ToLower(out), "username"), "Output should contain username")
		assert.True(t, strings.Contains(strings.ToLower(out), "password"), "Output should contain password")
		t.Log(out)
	})

	t.Run("CMD vault read database role demo-rw", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--info",
			"--unit-test",
			"--logical",
			"--path", "database/creds/demo-rw",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(strings.ToLower(out), "username"), "Output should contain username")
		assert.True(t, strings.Contains(strings.ToLower(out), "password"), "Output should contain password")
		t.Log(out)
	})

	t.Run("Connect to Database with Vault Credentials", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--info",
			"--unit-test",
			"--json",
			"--export=false",
			"--logical",
			"--path", "database/creds/demo-ro",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "vault read should not return an error: %s", err)
		connectVaultDBCredentials(t, out, pgContainer)
	})

	t.Run("Connect to Database with Vault Export Credentials", func(t *testing.T) {
		args := []string{
			"vault",
			"read",
			"--info",
			"--unit-test",
			"--export",
			"--json=false",
			"--logical",
			"--path", "database/creds/demo-ro",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "vault read should not return an error: %s", err)
		connectVaultDBExportCredentials(t, out, pgContainer)
	})
}

func connectVaultDBCredentials(t *testing.T, out string, pgContainer *dockertest.Resource) {
	t.Helper()
	// find the JSON part in the output
	jsonStart := strings.LastIndex(out, "{")
	require.GreaterOrEqual(t, jsonStart, 0, "Output should contain JSON")
	jsonStr := out[jsonStart:]
	// strip trailing log messages if any
	if jsonEnd := strings.LastIndex(jsonStr, "}"); jsonEnd > 0 {
		jsonStr = jsonStr[:jsonEnd+1]
	}

	var rawData map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &rawData), "Failed to unmarshal Vault JSON")

	data := rawData
	if nested, ok := rawData["data"].(map[string]any); ok {
		data = nested
	}
	dbUser, ok1 := data["username"].(string)
	dbPass, ok2 := data["password"].(string)
	require.True(t, ok1 && ok2, "Username or password not found in Vault response")

	pgHost, pgPort := common.GetContainerHostAndPort(pgContainer, "5432/tcp")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/demo?sslmode=disable", dbUser, dbPass, pgHost, pgPort)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to open database connection")
	defer func() { _ = db.Close() }()

	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one), "Failed to execute query with Vault credentials")
	assert.Equal(t, 1, one)
	t.Logf("Successfully connected to database with user: %s", dbUser)
}

func connectVaultDBExportCredentials(t *testing.T, out string, pgContainer *dockertest.Resource) {
	t.Helper()
	var dbUser, dbPass string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export USERNAME=") {
			dbUser = strings.Trim(strings.TrimPrefix(line, "export USERNAME="), "\"")
		}
		if strings.HasPrefix(line, "export PASSWORD=") {
			dbPass = strings.Trim(strings.TrimPrefix(line, "export PASSWORD="), "\"")
		}
	}
	require.NotEmpty(t, dbUser, "Username not found in Vault export response")
	require.NotEmpty(t, dbPass, "Password not found in Vault export response")

	pgHost, pgPort := common.GetContainerHostAndPort(pgContainer, "5432/tcp")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/demo?sslmode=disable", dbUser, dbPass, pgHost, pgPort)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to open database connection")
	defer func() { _ = db.Close() }()

	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one), "Failed to execute query with Vault export credentials")
	assert.Equal(t, 1, one)
	t.Logf("Successfully connected to database with exported user: %s", dbUser)
}
