package cmd

import (
	"fmt"
	"os"

	"github.com/tommi2day/pwcli/test"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
)

func TestVault(t *testing.T) {
	var err error
	var out = ""
	test.Testinit(t)
	if os.Getenv("SKIP_VAULT") != "" {
		t.Skip("Skip Vault Test in CI")
		return
	}
	vaultContainer, err := prepareVaultContainer()
	require.NoErrorf(t, err, "Ldap Server not available")
	require.NotNil(t, vaultContainer, "Prepare failed")
	defer common.DestroyDockerContainer(vaultContainer)

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
		require.NoErrorf(t, err, "get command should  not return an error: %s", err)
		assert.Contains(t, out, "Vault Write OK", "Output should not confirm success")
		t.Logf(out)
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
		t.Logf(out)
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
		t.Logf(out)
	})
	t.Run("CMD GetPassword Vault", func(t *testing.T) {
		args := []string{
			"get",
			"--method", "vault",
			"--info",
			"--unit-test",
			"--path", "secret/data/test",
			"--entry", "password",
			"--vault_addr", address,
			"--vault_token", rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm success")
		t.Logf(out)
	})
}
