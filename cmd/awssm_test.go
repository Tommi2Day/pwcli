package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/pwcli/test"
)

const awssmSecretPath = "pwcli-test/secret" //nolint:gosec // test fixture path, not a credential

func TestAWSSM(t *testing.T) {
	var err error
	var out = ""
	test.InitTestDirs()
	if os.Getenv("SKIP_AWSSM") != "" {
		t.Skip("Skip AWS Secrets Manager Test in CI")
		return
	}
	motoContainer, err := prepareMotoContainer()
	defer common.DestroyDockerContainer(motoContainer)
	require.NoErrorf(t, err, "Secrets Manager mock Server not available")
	require.NotNil(t, motoContainer, "Prepare failed")
	if err != nil || motoContainer == nil {
		t.Fatal("Secrets Manager mock server not available")
	}

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "abcdef")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "abcdefSecret")
	_ = os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
	_ = os.Setenv("SECRETSMANAGER_ENDPOINT", motoAddress)
	t.Logf("ADDR=%s", motoAddress)

	t.Run("CMD awssm write", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdWrite,
			flagInfo,
			flagUnitTest,
			flagPath, awssmSecretPath,
			flagAwssmEndpoint, motoAddress,
			"{\"password\": \"testpass\"}",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Write command should not return an error: %s", err)
		assert.Contains(t, out, "Secrets Manager Write OK", "Output should confirm success")
		t.Log(out)
	})

	t.Run("CMD awssm read", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdRead,
			flagDebug,
			flagUnitTest,
			flagPath, awssmSecretPath,
			flagAwssmEndpoint, motoAddress,
			entryPassword,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "read command should not return an error:%s", err)
		assert.Contains(t, out, "Secrets Manager Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		t.Log(out)
	})
	viper.Reset()

	t.Run("CMD awssm read json", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdRead,
			flagDebug,
			flagUnitTest,
			flagPath, awssmSecretPath,
			"--json",
			flagAwssmEndpoint, motoAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "read command should not return an error:%s", err)
		assert.Contains(t, out, "Secrets Manager Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		assert.True(t, strings.Contains(out, "{"), "Output should be json")
		t.Log(out)
	})
	viper.Reset()
	jsonOut = false

	t.Run("CMD awssm read export", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdRead,
			flagDebug,
			flagUnitTest,
			flagPath, awssmSecretPath,
			"--export",
			flagAwssmEndpoint, motoAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "read command should not return an error:%s", err)
		assert.True(t, strings.Contains(out, "export PASSWORD='testpass'"), "Output should be export format")
		t.Log(out)
	})
	viper.Reset()
	exportOut = false

	t.Run("CMD awssm read dotenv", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdRead,
			flagDebug,
			flagUnitTest,
			flagPath, awssmSecretPath,
			"--dotenv",
			flagAwssmEndpoint, motoAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "read command should not return an error:%s", err)
		assert.True(t, strings.Contains(out, "PASSWORD='testpass'"), "Output should be dotenv format")
		assert.False(t, strings.Contains(out, "export PASSWORD"), "Output should not contain export prefix")
		t.Log(out)
	})
	viper.Reset()
	dotenvOut = false

	t.Run("CMD awssm write with kms key", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdWrite,
			flagInfo,
			flagUnitTest,
			flagPath, awssmSecretPath + "-kms",
			flagAwssmEndpoint, motoAddress,
			"--awssm_kms_keyid", "alias/pwcli-awssm-test",
			"{\"password\": \"kmspass\"}",
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		require.NoErrorf(t, err, "Write command with kms key should not return an error: %s", err)
		assert.Contains(t, out, "Secrets Manager Write OK", "Output should confirm success")
	})
	viper.Reset()
	awssmKMSKeyID = ""
	_ = os.Unsetenv("SECRETSMANAGER_KMS_KEY_ID")

	t.Run("CMD awssm secrets", func(t *testing.T) {
		args := []string{
			typeAWSSM,
			cmdList,
			flagInfo,
			flagUnitTest,
			flagAwssmEndpoint, motoAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		require.NoErrorf(t, err, "list command should not return an error:%s", err)
		assert.Contains(t, out, awssmSecretPath, "Output should contain the written secret name")
	})
	viper.Reset()

	t.Run("CMD GetPassword AWSSM", func(t *testing.T) {
		args := []string{
			cmdGet,
			flagMethod, typeAWSSM,
			flagDebug,
			flagUnitTest,
			flagConfig, test.TestData + "/test_pwcli.yaml",
			flagPath, awssmSecretPath,
			"--entry", entryPassword,
			flagAwssmEndpoint, motoAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm success")
		t.Log(out)
	})
}
