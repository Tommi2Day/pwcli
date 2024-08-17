package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/tommi2day/pwcli/test"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
)

func TestKMS(t *testing.T) {
	if os.Getenv("SKIP_KMS") != "" {
		t.Skip("Skipping KMS testing in CI environment")
	}
	var err error
	var out = ""
	test.Testinit(t)

	kmsapp := "test_kms_file"
	testdata := test.TestData
	kmspc := pwlib.NewConfig(kmsapp, testdata, testdata, kmsapp, typeKMS)

	err = os.Chdir(test.TestDir)
	require.NoErrorf(t, err, "ChDir failed")
	filename := kmspc.PlainTextFile
	_ = os.Remove(filename)
	//nolint gosec
	err = os.WriteFile(filename, []byte(plain), 0644)
	require.NoErrorf(t, err, "Create testdata failed")

	var kmsClient *kms.Client
	kmsContainer, err := prepareKmsContainer()
	require.NoErrorf(t, err, "KMS Server not available")
	require.NotNil(t, kmsContainer, "Prepare failed")
	defer common.DestroyDockerContainer(kmsContainer)

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "abcdef")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "abcdefSecret")
	_ = os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
	_ = os.Setenv("KMS_ENDPOINT", kmsAddress)

	kmsClient = pwlib.ConnectToKMS()
	require.NotNil(t, kmsClient, "Connect to KMS failed")
	keyID := ""
	alias := fmt.Sprintf("alias/%s", kmsapp)
	if kmsClient == nil {
		t.Fatal("Connect to KMS failed")
	}
	t.Run("Create KMS Key", func(t *testing.T) {
		keyout, err := pwlib.GenKMSKey(kmsClient, "", fmt.Sprintf("Key for %s", kmsapp), map[string]string{"app": kmsapp})
		require.NoErrorf(t, err, "CreateKMSKey failed:%s", err)
		require.NotNil(t, keyout, "CreateKMSKey response empty")
		if keyout != nil {
			keyID, _ = pwlib.GetKMSKeyIDs(keyout.KeyMetadata)
			_, err = pwlib.CreateKMSAlias(kmsClient, alias, keyID)
			require.NoErrorf(t, err, "CreateKMSAlias failed:%s", err)
		}
	})
	if keyID == "" {
		t.Fatal("CreateKMSKey failed")
		return
	}

	t.Run("CMD Encrypt KMS", func(t *testing.T) {
		args := []string{
			"encrypt",
			"--app", kmsapp,
			"-D", testdata,
			"-K", testdata,
			"--kms_keyid", keyID,
			"--method", typeKMS,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, pc.CryptedFile, "Crypted file '%s' not found", pc.CryptedFile)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Log(out)
	})
	viper.Set("kms_keyid", "")
	t.Run("CMD list KMS with alias", func(t *testing.T) {
		args := []string{
			"list",
			"--app", kmsapp,
			"-D", testdata,
			"-K", testdata,
			"--method", typeKMS,
			"--kms_keyid", alias,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "list command should not return an error:%s", err)
		assert.Contains(t, out, "List returned 10 lines", "Output should lines of plainfile")
		t.Log(out)
	})

	t.Run("CMD get KMS with Key Env", func(t *testing.T) {
		_ = os.Setenv("KMS_KEYID", keyID)
		args := []string{
			"get",
			"--app", kmsapp,
			"--method", typeKMS,
			"-D", testdata,
			"-K", testdata,
			"--info",
			"--unit-test",
			"--system", "test",
			"--user", "testuser",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm match")
		assert.Contains(t, out, "'testpass'", "Output should return correct match")
		t.Log(out)
	})
}
