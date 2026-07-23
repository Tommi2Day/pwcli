package cmd

import (
	"fmt"
	"os"
	"path"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/pwcli/test"
)

func TestEncryptDecryptTypes(t *testing.T) {
	viper.Reset()
	test.InitTestDirs()
	_ = os.Mkdir(test.TestData, 0700)
	err := os.Chdir(test.TestDir)
	require.NoError(t, err)

	const (
		testapp   = "test_crypt"
		testpass  = "testpass"
		plaintext = "secret data\nmultiple lines\n"
	)

	// rsa is already tested in TestCLI, but we'll include it for completeness if we want,
	// but here we focus on age and gpg which are new.
	// Note: ecdsa does not support encryption.
	types := []struct {
		kt     string
		method string
	}{
		{defaultKeyType, "go"},
		{defaultKeyType, "openssl"},
		{typeAGE, typeAGE},
		{typeGPG, typeGPG},
	}

	for _, tc := range types {
		t.Run(fmt.Sprintf("Type_%s_Method_%s", tc.kt, tc.method), func(t *testing.T) {
			appname := fmt.Sprintf("%s_%s", testapp, tc.kt)
			// 1. Generate Key
			genArgs := []string{
				cmdGenkey,
				flagType, tc.kt,
				flagMethod, tc.method,
				flagKeypass, testpass,
				flagApp, appname,
				flagDatadir, test.TestData,
				flagKeydir, test.TestData,
				flagUnitTest,
				flagDebug,
			}
			log.Debugf("Running genkey for %s\n%v", tc.kt, genArgs)
			_, err := common.CmdRun(RootCmd, genArgs)
			require.NoError(t, err, "genkey failed for %s", tc.kt)

			// 2. Prepare plaintext file
			plainFile := path.Join(test.TestData, appname+".plain")
			err = common.WriteStringToFile(plainFile, plaintext)
			require.NoError(t, err)

			// 3. Encrypt
			encArgs := []string{
				cmdEncrypt,
				flagDebug,
				flagMethod, tc.method,
				flagKeypass, testpass,
				flagApp, appname,
				flagDatadir, test.TestData,
				flagKeydir, test.TestData,
				flagPlaintext, plainFile,
				flagUnitTest,
				flagDebug,
			}
			out, err := common.CmdRun(RootCmd, encArgs)
			require.NoError(t, err, "encrypt failed for method %s: %s", tc.method, out)
			// assert.Contains(t, out, "DONE")

			// 4. Decrypt
			decPlainFile := path.Join(test.TestData, appname+".decrypted")
			decArgs := []string{
				"decrypt",
				flagMethod, tc.method,
				flagKeypass, testpass,
				flagApp, appname,
				flagDatadir, test.TestData,
				flagKeydir, test.TestData,
				flagPlaintext, decPlainFile,
				flagUnitTest,
				flagDebug,
			}
			out, err = common.CmdRun(RootCmd, decArgs)
			require.NoError(t, err, "decrypt failed for method %s: %s", tc.method, out)
			// assert.Contains(t, out, "DONE")

			// 5. Verify content
			decryptedContent, err := common.ReadFileToString(decPlainFile)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decryptedContent)
		})
	}
}
