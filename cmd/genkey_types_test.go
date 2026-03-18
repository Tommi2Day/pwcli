package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
	"github.com/tommi2day/pwcli/test"
)

func TestGenKeyTypes(t *testing.T) {
	testapp := "test_genkey"
	test.InitTestDirs()
	_ = os.Mkdir(test.TestData, 0700)

	// Set up environment similar to TestCLI
	err := os.Chdir(test.TestDir)
	require.NoError(t, err)

	types := []string{"rsa", "ecdsa", "age", "gpg"}

	for _, kt := range types {
		t.Run("genkey type "+kt, func(t *testing.T) {
			// viper.Reset()
			appname := testapp + "_" + kt
			// NewConfig sets up paths for keys based on method,
			// but genkey uses pc.PrivateKeyFile and pc.PubKeyFile which are initialized here.
			// For age and gpg, pwlib.NewConfig will return different extensions if we pass the method.

			// localPC := pwlib.NewConfig(appname, test.TestData, test.TestData, kt, method)

			args := []string{
				"genkey",
				"--type", kt,
				"--keypass", kp,
				"--app", appname,
				"--datadir", test.TestData,
				"--keydir", test.TestData,
				"--unit-test",
			}

			out, err := common.CmdRun(RootCmd, args)
			require.NoErrorf(t, err, "genkey --type %s failed: %v", kt, err)
			assert.Contains(t, out, "DONE")

			// We need to know what files were actually generated.
			// Since genkey uses pc (global), let's check pc.
			assert.FileExists(t, pc.PrivateKeyFile, "Private key file for %s not found: %s", kt, pc.PrivateKeyFile)
			assert.FileExists(t, pc.PubKeyFile, "Public key file for %s not found: %s", kt, pc.PubKeyFile)

			// Verify key type from file
			detectedType, err := pwlib.GetKeyTypeFromFile(pc.PrivateKeyFile)
			assert.NoError(t, err)

			expectedType := pwlib.KeyTypeUnknown
			switch kt {
			case "rsa":
				expectedType = pwlib.KeyTypeRSA
			case "ecdsa":
				expectedType = pwlib.KeyTypeECDSA
			case "age":
				expectedType = pwlib.KeyTypeAGE
			case "gpg":
				expectedType = pwlib.KeyTypeGPG
			}
			assert.Equal(t, expectedType, detectedType, "Detected key type mismatch for %s", kt)
		})
	}
}
