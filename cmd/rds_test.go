package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
)

const (
	flagAWSProfile  = "--aws_profile"
	flagRDSRegion   = "--rds_region"
	rdsTestEndpoint = "mydb.abc123.eu-central-1.rds.amazonaws.com"
	rdsTestUser     = "dbuser"
	rdsTestProfile  = "pwcli-rds"
	rdsTestKeyID    = "AKIAPWCLIRDSTEST"
)

// setupRDSAWSConfig writes an isolated AWS shared config with a static credential profile
func setupRDSAWSConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config")
	content := "[profile " + rdsTestProfile + "]\n" +
		"aws_access_key_id = " + rdsTestKeyID + "\n" +
		"aws_secret_access_key = pwcliRDSTestSecret\n" +
		"region = eu-west-1\n"
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0600))
	t.Setenv("AWS_CONFIG_FILE", cfgFile)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	t.Setenv("RDS_REGION", "")
}

// resetRDSState drops flag state leaking between subtests, incl. RDS_REGION set by --rds_region
func resetRDSState() {
	viper.Reset()
	rdsRegion = ""
	_ = os.Setenv("RDS_REGION", "")
	method = defaultType
	awsProfile = ""
	if f := RootCmd.PersistentFlags().Lookup(keyAWSProfile); f != nil {
		f.Changed = false
	}
}

func TestRDSToken(t *testing.T) {
	setupRDSAWSConfig(t)
	t.Cleanup(resetRDSState)

	t.Run("token with profile", func(t *testing.T) {
		resetRDSState()
		args := []string{typeRDS, cmdToken, flagInfo, flagUnitTest, flagNoPrompt,
			flagAWSProfile, rdsTestProfile, "--endpoint", rdsTestEndpoint, flagUser, rdsTestUser}
		out, err := common.CmdRun(RootCmd, args)
		t.Log(out)
		require.NoError(t, err)
		assert.Contains(t, out, "RDS auth token successfully generated")
	})

	t.Run("token missing user", func(t *testing.T) {
		resetRDSState()
		args := []string{typeRDS, cmdToken, flagUnitTest, flagNoPrompt,
			flagAWSProfile, rdsTestProfile, "--endpoint", rdsTestEndpoint, flagUser, ""}
		_, err := common.CmdRun(RootCmd, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint and user")
	})

	t.Run("token unknown profile", func(t *testing.T) {
		resetRDSState()
		args := []string{typeRDS, cmdToken, flagUnitTest, flagNoPrompt,
			flagAWSProfile, "pwcli-does-not-exist", "--endpoint", rdsTestEndpoint, flagUser, rdsTestUser}
		_, err := common.CmdRun(RootCmd, args)
		require.Error(t, err, "unknown profile must return an error instead of terminating")
	})
}

func TestGetPasswordRDS(t *testing.T) {
	setupRDSAWSConfig(t)
	t.Cleanup(resetRDSState)

	t.Run("get with method rds", func(t *testing.T) {
		resetRDSState()
		args := []string{cmdGet, flagMethod, typeRDS, flagInfo, flagUnitTest, flagNoPrompt,
			flagAWSProfile, rdsTestProfile, flagRDSRegion, "eu-central-1",
			flagSystem, rdsTestEndpoint + ":3306", flagUser, rdsTestUser}
		out, err := common.CmdRun(RootCmd, args)
		t.Log(out)
		require.NoError(t, err)
		assert.Contains(t, out, "Found matching entry: '"+rdsTestEndpoint+":3306?Action=connect")
		assert.Contains(t, out, "DBUser="+rdsTestUser)
		assert.Contains(t, out, rdsTestKeyID+"%2F")
		assert.Contains(t, out, "%2Feu-central-1%2Frds-db%2F", "region flag should override profile region")
	})

	t.Run("get with method rds default port and profile region", func(t *testing.T) {
		resetRDSState()
		args := []string{cmdGet, flagMethod, typeRDS, flagInfo, flagUnitTest, flagNoPrompt,
			flagAWSProfile, rdsTestProfile, flagSystem, rdsTestEndpoint, flagUser, rdsTestUser}
		out, err := common.CmdRun(RootCmd, args)
		t.Log(out)
		require.NoError(t, err)
		assert.Contains(t, out, rdsTestEndpoint+":5432?Action=connect")
		assert.Contains(t, out, "%2Feu-west-1%2Frds-db%2F")
	})

	t.Run("get with method rds without endpoint", func(t *testing.T) {
		resetRDSState()
		args := []string{cmdGet, flagMethod, typeRDS, flagUnitTest, flagNoPrompt,
			flagAWSProfile, rdsTestProfile, flagSystem, "", "--db", "", flagUser, rdsTestUser}
		_, err := common.CmdRun(RootCmd, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method rds needs parameter system/db")
	})
}
