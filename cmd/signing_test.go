package cmd

import (
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
	"github.com/tommi2day/pwcli/test"
)

func TestSignVerify(t *testing.T) {
	viper.Reset()
	var err error
	var out = ""
	const testapp = "test_sign"
	const kp = "testpass"
	const plainContent = "This is a test message to sign"

	test.InitTestDirs()
	_ = os.Mkdir(test.TestData, 0700)
	configFile := path.Join(test.TestData, testapp+".yaml")
	err = os.Chdir(test.TestDir)
	require.NoErrorf(t, err, "ChDir failed")

	pc = pwlib.NewConfig(testapp, test.TestData, test.TestData, kp, typeGO)
	plaintextFile := pc.PlainTextFile
	_ = os.Remove(plaintextFile)
	err = common.WriteStringToFile(plaintextFile, plainContent)
	require.NoErrorf(t, err, "Create testdata failed")

	t.Run("Generate Keypair for Signing", func(t *testing.T) {
		args := []string{
			"genkey",
			"--keypass", kp,
			"--method", typeGO,
			"--app", testapp,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
			"--type", pwlib.KeyTypeRSA,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Generate command failed: %s", err)
		assert.FileExists(t, pc.PrivateKeyFile)
		assert.FileExists(t, pc.PubKeyFile)
	})

	t.Run("Sign File", func(t *testing.T) {
		args := []string{
			"sign",
			"--app", testapp,
			"--method", typeGO,
			"--keypass", kp,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Sign command failed: %s", err)
		assert.Contains(t, out, "DONE")
		assert.FileExists(t, pc.SignatureFile)
	})

	t.Run("Verify File Valid", func(t *testing.T) {
		args := []string{
			"verify",
			"--app", testapp,
			"--method", typeGO,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Verify command failed: %s", err)
		assert.Contains(t, out, "VALID")
	})

	t.Run("Verify File Invalid", func(t *testing.T) {
		// tamper with plaintext
		err = common.WriteStringToFile(plaintextFile, plainContent+" tampered")
		require.NoError(t, err)

		args := []string{
			"verify",
			"--app", testapp,
			"--method", typeGO,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		// It might return error if verification fails, or just print INVALID
		// In my implementation it prints INVALID and returns nil error if VerifyFile returns false but no err.
		// Wait, if VerifyFile returns false, it means signature is invalid.
		assert.Contains(t, out, "INVALID")
	})

	t.Run("GPG Sign and Verify", func(t *testing.T) {
		const gpgApp = "test_gpg"
		gpgPc := pwlib.NewConfig(gpgApp, test.TestData, test.TestData, kp, typeGPG)
		_ = common.WriteStringToFile(gpgPc.PlainTextFile, plainContent)

		// Generate GPG key
		args := []string{
			"genkey",
			"--keypass", kp,
			"--method", typeGPG,
			"--app", gpgApp,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
			"--type", pwlib.KeyTypeGPG,
		}
		_, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)

		// Sign
		args = []string{
			"sign",
			"--app", gpgApp,
			"--method", typeGPG,
			"--keypass", kp,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, "DONE")

		// Verify
		args = []string{
			"verify",
			"--app", gpgApp,
			"--method", typeGPG,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, "VALID")

		// Cleanup GPG
		_ = os.Remove(gpgPc.PrivateKeyFile)
		_ = os.Remove(gpgPc.PubKeyFile)
		_ = os.Remove(gpgPc.PlainTextFile)
		_ = os.Remove(gpgPc.SignatureFile)
	})

	// Cleanup
	_ = os.Remove(configFile)
	_ = os.Remove(pc.PrivateKeyFile)
	_ = os.Remove(pc.PubKeyFile)
	_ = os.Remove(pc.PlainTextFile)
	_ = os.Remove(pc.SignatureFile)
}
