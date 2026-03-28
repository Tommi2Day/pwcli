package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
	"github.com/tommi2day/pwcli/test"
)

func TestGopassCLI(t *testing.T) {
	viper.Reset()
	test.InitTestDirs()
	_ = os.MkdirAll(test.TestData, 0700)

	storeDir := filepath.Join(test.TestData, "gopass-store")
	require.NoError(t, os.MkdirAll(storeDir, 0700))

	// Generate an age key pair for the test store
	identity, _, err := pwlib.CreateAgeIdentity()
	require.NoErrorf(t, err, "CreateAgeIdentity failed")

	pubKeyFile := filepath.Join(test.TestData, "gopass-test.pub")
	privKeyFile := filepath.Join(test.TestData, "gopass-test.age")
	err = pwlib.ExportAgeKeyPair(identity, pubKeyFile, privKeyFile)
	require.NoErrorf(t, err, "ExportAgeKeyPair failed")

	// Seed store with the age recipients marker so crypto auto-detection works
	pubKeyContent, err := common.ReadFileToString(pubKeyFile)
	require.NoErrorf(t, err, "failed to read public key file")
	recipientsFile := filepath.Join(storeDir, ".age-recipients")
	err = os.WriteFile(recipientsFile, []byte(pubKeyContent+"\n"), 0600)
	require.NoErrorf(t, err, "failed to write .age-recipients")

	const secretName = "test/mypassword"
	const secretContent = "mysecretpassword"

	t.Run("gopass write", func(t *testing.T) {
		args := []string{
			"gopass", "write", secretName,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--key-file", pubKeyFile,
			"--content", secretContent,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass write failed: %s\n%s", err, out)
		assert.Contains(t, out, "written")
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass list", func(t *testing.T) {
		args := []string{
			"gopass", "list",
			"--store-dir", storeDir,
			"--crypto", "age",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass list failed: %s\n%s", err, out)
		assert.Contains(t, out, secretName)
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass read", func(t *testing.T) {
		args := []string{
			"gopass", "read", secretName,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--key-file", privKeyFile,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass read failed: %s\n%s", err, out)
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass read raw", func(t *testing.T) {
		args := []string{
			"gopass", "read", secretName,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--key-file", privKeyFile,
			"--raw",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass read --raw failed: %s\n%s", err, out)
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass recipients list", func(t *testing.T) {
		args := []string{
			"gopass", "recipients", "list",
			"--store-dir", storeDir,
			"--crypto", "age",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass recipients list failed: %s\n%s", err, out)
		assert.Contains(t, out, "age1")
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass recipients add", func(t *testing.T) {
		const newRecipient = "age1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs"
		args := []string{
			"gopass", "recipients", "add", newRecipient,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass recipients add failed: %s\n%s", err, out)
		assert.Contains(t, out, "recipient added")

		data, readErr := os.ReadFile(recipientsFile)
		require.NoErrorf(t, readErr, "failed to read recipients file after add")
		assert.Contains(t, string(data), newRecipient)
		t.Log(out)
	})

	identityDir := filepath.Join(test.TestData, "gopass-identities")
	gopassIdentityDir = "" // reset between tests

	createIdentityDir := filepath.Join(test.TestData, "gopass-created-identities")

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass identity create age", func(t *testing.T) {
		args := []string{
			"gopass", "identity", "create", "newage",
			"--crypto", "age",
			"--identity-dir", createIdentityDir,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity create age failed: %s\n%s", err, out)
		assert.Contains(t, out, "newage")
		assert.Contains(t, out, "age")
		assert.FileExists(t, filepath.Join(createIdentityDir, "newage.key"))
		assert.FileExists(t, filepath.Join(createIdentityDir, "newage.pub"))
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass identity create age with add-recipient", func(t *testing.T) {
		args := []string{
			"gopass", "identity", "create", "newage2",
			"--crypto", "age",
			"--identity-dir", createIdentityDir,
			"--store-dir", storeDir,
			"--add-recipient",
			"--unit-test",
		}
		recipientsBefore, _ := os.ReadFile(filepath.Join(storeDir, ".age-recipients"))
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity create age+add-recipient failed: %s\n%s", err, out)
		assert.Contains(t, out, "added to recipients")
		assert.FileExists(t, filepath.Join(createIdentityDir, "newage2.key"))
		recipientsAfter, _ := os.ReadFile(filepath.Join(storeDir, ".age-recipients"))
		assert.Greater(t, len(recipientsAfter), len(recipientsBefore), "recipients file should have grown")
		t.Log(out)
	})

	const encPassphrase = "testpassphrase"
	encIdentityDir := filepath.Join(test.TestData, "gopass-enc-identities")
	encStoreDir := filepath.Join(test.TestData, "gopass-enc-store")

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass identity create age with passphrase", func(t *testing.T) {
		require.NoError(t, os.MkdirAll(encStoreDir, 0700))
		args := []string{
			"gopass", "identity", "create", "newage-enc",
			"--crypto", "age",
			"--identity-dir", encIdentityDir,
			"--store-dir", encStoreDir,
			"--passphrase", encPassphrase,
			"--add-recipient",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity create age+passphrase failed: %s\n%s", err, out)
		assert.Contains(t, out, "newage-enc")
		assert.FileExists(t, filepath.Join(encIdentityDir, "newage-enc.key"))
		assert.FileExists(t, filepath.Join(encIdentityDir, "newage-enc.pub"))
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass write to enc-identity store", func(t *testing.T) {
		args := []string{
			"gopass", "write", "enc/secret",
			"--store-dir", encStoreDir,
			"--crypto", "age",
			"--key-file", filepath.Join(encIdentityDir, "newage-enc.pub"),
			"--content", "encryptedsecret",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass write enc-identity store failed: %s\n%s", err, out)
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	gopassKeyFile = ""
	gopassStoreDir = ""
	t.Run("gopass read prompts for passphrase when identity is encrypted", func(t *testing.T) {
		// Inject the passphrase via a pipe so PromptPassword reads it from common.InputReader
		pr, pw, pipeErr := os.Pipe()
		require.NoError(t, pipeErr)
		_, _ = pw.WriteString(encPassphrase + "\n")
		_ = pw.Close()
		common.InputReader = pr
		defer func() { common.InputReader = os.Stdin }()
		args := []string{
			"gopass", "read", "enc/secret",
			"--store-dir", encStoreDir,
			"--crypto", "age",
			"--identity-dir", encIdentityDir,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass read with prompted passphrase failed: %s\n%s", err, out)
		assert.Contains(t, out, "encryptedsecret")
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	gopassKeyFile = ""
	gopassStoreDir = ""
	noPromptFlag = false
	t.Run("gopass read --no-prompt errors instead of prompting", func(t *testing.T) {
		args := []string{
			"gopass", "read", "enc/secret",
			"--store-dir", encStoreDir,
			"--crypto", "age",
			"--identity-dir", encIdentityDir,
			"--no-prompt",
			"--unit-test",
		}
		_, err := common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "gopass read --no-prompt should return an error")
		assert.Contains(t, err.Error(), "no-prompt")
	})

	viper.Reset()
	gopassIdentityDir = ""
	gopassKeyFile = ""
	noPromptFlag = false
	t.Run("gopass identity create gpg", func(t *testing.T) {
		args := []string{
			"gopass", "identity", "create", "newgpg",
			"--crypto", "gpg",
			"--identity-dir", createIdentityDir,
			"--name", "Test User",
			"--email", "test@example.com",
			"--passphrase", "testpass",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity create gpg failed: %s\n%s", err, out)
		assert.Contains(t, out, "newgpg")
		assert.Contains(t, out, "gpg")
		assert.FileExists(t, filepath.Join(createIdentityDir, "newgpg.key"))
		assert.FileExists(t, filepath.Join(createIdentityDir, "newgpg.pub"))
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass identity add", func(t *testing.T) {
		args := []string{
			"gopass", "identity", "add", "mykey", privKeyFile,
			"--identity-dir", identityDir,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity add failed: %s\n%s", err, out)
		assert.Contains(t, out, "mykey")
		assert.FileExists(t, filepath.Join(identityDir, "mykey.key"))
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass read auto-detect via GOPASS_IDENTITY_DIR", func(t *testing.T) {
		_ = os.Setenv("GOPASS_IDENTITY_DIR", identityDir)
		defer func() { _ = os.Unsetenv("GOPASS_IDENTITY_DIR") }()
		args := []string{
			"gopass", "read", secretName,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass read GOPASS_IDENTITY_DIR failed: %s\n%s", err, out)
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass read auto-detect via config age.identity_dir", func(t *testing.T) {
		// Write a gopass config that sets age.identity_dir.
		cfgContent := fmt.Sprintf(
			"root:\n  path: %s\n  crypto: age\nmounts: {}\nage:\n  identity_dir: %s\n",
			storeDir, identityDir,
		)
		cfgPath := filepath.Join(test.TestData, "gopass-id-config.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))
		_ = os.Setenv("GOPASS_CONFIG", cfgPath)
		defer func() { _ = os.Unsetenv("GOPASS_CONFIG") }()
		args := []string{
			"gopass", "read", secretName,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass read config age.identity_dir failed: %s\n%s", err, out)
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	gopassIdentityDir = ""
	t.Run("gopass read auto-detect via config path sibling", func(t *testing.T) {
		// Config sits inside identityDir itself → sibling "identities/" would be
		// identityDir/identities/. Instead place the config one level up so that
		// the sibling of the config file IS identityDir.
		// Layout: test.TestData/sibling-test/config  →  test.TestData/sibling-test/identities/
		siblingBase := filepath.Join(test.TestData, "sibling-test")
		siblingIdentities := filepath.Join(siblingBase, "identities")
		require.NoError(t, os.MkdirAll(siblingIdentities, 0700))
		// Copy the private key into the sibling identities dir.
		keyData, err := os.ReadFile(filepath.Clean(privKeyFile))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Clean(filepath.Join(siblingIdentities, "mykey.key")), keyData, 0600)) //nolint:gosec
		// Write a minimal config next to it (no age.identity_dir field).
		cfgContent := fmt.Sprintf("root:\n  path: %s\n  crypto: age\nmounts: {}\n", storeDir)
		cfgPath := filepath.Join(siblingBase, "config")
		require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))
		_ = os.Setenv("GOPASS_CONFIG", cfgPath)
		defer func() { _ = os.Unsetenv("GOPASS_CONFIG") }()
		args := []string{
			"gopass", "read", secretName,
			"--store-dir", storeDir,
			"--crypto", "age",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass read sibling identity dir failed: %s\n%s", err, out)
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass identity list", func(t *testing.T) {
		args := []string{
			"gopass", "identity", "list",
			"--identity-dir", identityDir,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity list failed: %s\n%s", err, out)
		assert.Contains(t, out, "mykey.key")
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass identity list empty dir", func(t *testing.T) {
		emptyDir := filepath.Join(test.TestData, "gopass-identities-empty")
		args := []string{
			"gopass", "identity", "list",
			"--identity-dir", emptyDir,
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass identity list on missing dir should not error: %s\n%s", err, out)
		t.Log(out)
	})

	viper.Reset()
	t.Run("gopass stores", func(t *testing.T) {
		configContent := fmt.Sprintf("root:\n  path: %s\n  crypto: age\nmounts: {}\n", storeDir)
		configPath := filepath.Join(test.TestData, "gopass-config.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0600)
		require.NoErrorf(t, err, "failed to write temp gopass config")
		_ = os.Setenv("GOPASS_CONFIG", configPath)
		defer func() { _ = os.Unsetenv("GOPASS_CONFIG") }()

		args := []string{
			"gopass", "stores",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "gopass stores failed: %s\n%s", err, out)
		assert.Contains(t, out, "root:")
		assert.Contains(t, out, storeDir)
		t.Log(out)
	})

	viper.Reset()
	gopassStoreDir = ""
	gopassKeyFile = ""
	t.Run("get --method gopass password field", func(t *testing.T) {
		args := []string{
			"get",
			"--method", "gopass",
			"--path", secretName,
			"--store-dir", storeDir,
			"--key-file", privKeyFile,
			"--info",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get gopass failed: %s\n%s", err, out)
		assert.Contains(t, out, "Found matching entry")
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	gopassStoreDir = ""
	gopassKeyFile = ""
	t.Run("get --method gopass explicit entry field", func(t *testing.T) {
		args := []string{
			"get",
			"--method", "gopass",
			"--path", secretName,
			"--entry", "password",
			"--store-dir", storeDir,
			"--key-file", privKeyFile,
			"--info",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get gopass with entry failed: %s\n%s", err, out)
		assert.Contains(t, out, "Found matching entry")
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	gopassStoreDir = ""
	gopassKeyFile = ""
	gopassIdentityDir = ""
	t.Run("get --method gopass auto-detect via GOPASS_IDENTITY_DIR", func(t *testing.T) {
		_ = os.Setenv("GOPASS_IDENTITY_DIR", identityDir)
		defer func() { _ = os.Unsetenv("GOPASS_IDENTITY_DIR") }()
		args := []string{
			"get",
			"--method", "gopass",
			"--path", secretName,
			"--store-dir", storeDir,
			"--info",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get gopass auto-detect failed: %s\n%s", err, out)
		assert.Contains(t, out, "Found matching entry")
		assert.Contains(t, out, secretContent)
		t.Log(out)
	})

	viper.Reset()
	gopassStoreDir = ""
	gopassKeyFile = ""
	t.Run("get --method gopass missing path", func(t *testing.T) {
		// Pass --path "" explicitly so cobra does not inherit the path from a prior run.
		args := []string{
			"get",
			"--method", "gopass",
			"--path", "",
			"--store-dir", storeDir,
			"--key-file", privKeyFile,
			"--unit-test",
		}
		_, err := common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "get gopass without --path should error")
		assert.Contains(t, err.Error(), "--path")
	})

	viper.Reset()
	gopassStoreDir = ""
	gopassKeyFile = ""
	t.Run("gopass list help hides global flags", func(t *testing.T) {
		args := []string{
			"gopass", "list",
			"--help",
			"--unit-test",
		}
		out, err := common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "help should not return an error: %s", err)
		assert.Contains(t, out, "--store-dir")
		assert.NotContains(t, out, "--datadir")
		t.Log(out)
	})
}
