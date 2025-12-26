package cmd

import (
	"fmt"
	"os"
	"testing"

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
	vaultContainer, err := prepareVaultContainer()
	defer common.DestroyDockerContainer(vaultContainer)
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
			"--mount", "secret/",
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
			"--mount", "secret/",
			"--path", "test",
			"--vault_addr", address,
			"--vault_token", rootToken,
			"password",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		t.Log(out)
	})
	t.Run("CMD vault list", func(t *testing.T) {
		args := []string{
			"vault",
			"list",
			"--info",
			"--unit-test",
			"--mount", "secret/",
			"--path", "/",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault List returned", "Output should confirm success")
		t.Log(out)
	})
	t.Run("CMD vault list error", func(t *testing.T) {
		args := []string{
			"vault",
			"list",
			"--info",
			"--unit-test",
			"--mount", "secret/",
			"--path", "/dummy",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "list command should  return an error")
		if err != nil {
			t.Log(err)
			assert.Contains(t, err.Error(), "no Entries returned", "Output should return no entries")
		}
		t.Log(out)
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
}
