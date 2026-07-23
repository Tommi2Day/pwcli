package cmd

import (
	"testing"

	"github.com/tommi2day/gomodules/pwlib"

	"golang.org/x/crypto/bcrypt"

	"github.com/tommi2day/gomodules/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hashPassword = "testHashPassword"
const hashUsername = "testHashUsername"
const testBcrypt = "$2a$10$Y3xlpzHMnNyZXm.rnIGqouf9NpPP.OCtB6FakJC3nK/Z1CYmC3Amq"
const testMD5 = "{MD5}ebcd5bc0483385f278b814600272d794"
const testSSHA = "{SSHA}r3myNFUMmkpxkaJ9EIr071i9x+1MqPgS"
const testBasic = "dGVzdEhhc2hVc2VybmFtZTp0ZXN0SGFzaFBhc3N3b3Jk"
const testArgon2 = "$argon2id$v=19$m=65536,t=3,p=4$yVSLalsV0ZyoyByE5IQDVg$V14dRnxoKArosnameO3QdnFstLMbGvqHhJsUbZ9UQcI"

func TestHash(t *testing.T) {
	var out string
	var err error
	t.Run("TestHashSSHACompare", func(t *testing.T) {
		var hash string
		hash, err = doSSHA(hashPassword, pwlib.SSHAPrefix)
		require.NoErrorf(t, err, "hash ssha command should  not return an error:%s", err)
		enc := pwlib.SSHAEncoder{}
		m := enc.Matches([]byte(hash), []byte(hashPassword))
		require.True(t, m, "ssha compare should  be true")
	})
	t.Run("TestHashSSHA", func(t *testing.T) {
		args := []string{
			cmdHash,
			"ssha",
			flagPassword, hashPassword,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash ssha command should  not return an error:%s", err)
		assert.Contains(t, out, pwlib.SSHAPrefix, "Output should contain SSHA header")
		t.Log(out)
	})
	t.Run("TestHashSSHAMatch", func(t *testing.T) {
		args := []string{
			cmdHash,
			"ssha",
			flagPassword, hashPassword,
			flagTest, testSSHA,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash ssha command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Log(out)
	})
	_ = md5Cmd.Flags().Set(testID, "")
	t.Run("TestHashMD5", func(t *testing.T) {
		args := []string{
			cmdHash,
			mMD5,
			flagUsername, hashUsername,
			flagPassword, hashPassword,
			"--prefix", mMD5,
			"--test=",
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash md5 command should  not return an error:%s", err)
		assert.Contains(t, out, "md5ebcd", "Output should contain MD5 prefix")
		t.Log(out)
	})
	t.Run("TestHashMD5Match", func(t *testing.T) {
		args := []string{
			cmdHash,
			mMD5,
			flagUsername, hashUsername,
			flagPassword, hashPassword,
			"--prefix", mMD5,
			flagTest, testMD5,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash md5 command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Log(out)
	})
	t.Run("TestHashScram", func(t *testing.T) {
		args := []string{
			cmdHash,
			"scram",
			flagUsername, hashUsername,
			flagPassword, hashPassword,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash scram command should  not return an error:%s", err)
		assert.Contains(t, out, "SCRAM-SHA-256$4096:", "Output should contain SCRAM-SHA-256 header")
		t.Log(out)
	})
	t.Run("TestHashBcryptCompare", func(t *testing.T) {
		var hash string
		hash, err = doBcrypt(hashPassword)
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(hashPassword))
		require.NoErrorf(t, err, "bcrypt compare should  not return an error:%s", err)
	})

	_ = hashCmd.Flags().Set(testID, "")
	t.Run("TestHashBcrypt", func(t *testing.T) {
		args := []string{
			cmdHash,
			"bcrypt",
			flagPassword, hashPassword,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash bcrypt command should  not return an error:%s", err)
		assert.Contains(t, out, "$2a$", "Output should contain bcrypt header")
		t.Log(out)
	})

	t.Run("TestHashBcryptMatch", func(t *testing.T) {
		args := []string{
			cmdHash,
			"bcrypt",
			flagPassword, hashPassword,
			flagTest, testBcrypt,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash bcrypt command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Log(out)
	})

	t.Run("TestHashBasic", func(t *testing.T) {
		args := []string{
			cmdHash,
			"basic",
			flagUsername, hashUsername,
			flagPassword, hashPassword,
			"--prefix", "Authorization: Basic ",
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash basic command should  not return an error:%s", err)
		assert.Contains(t, out, "Basic", "Output should contain Basic header")
		assert.Contains(t, out, testBasic, "Output not matches")
		t.Log(out)
	})
	t.Run("TestHashBasicMatch", func(t *testing.T) {
		args := []string{
			cmdHash,
			"basic",
			flagUsername, hashUsername,
			flagPassword, hashPassword,
			flagTest, testBasic,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash basic command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Log(out)
	})

	_ = hashCmd.Flags().Set(testID, "")
	t.Run("TestHashArgon2", func(t *testing.T) {
		args := []string{
			cmdHash,
			"argon2",
			flagPassword, hashPassword,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash argon2 command should  not return an error:%s", err)
		assert.Contains(t, out, "$argon2id$", "Output should contain argon header")
		t.Log(out)
	})

	t.Run("TestHashArgon2tMatch", func(t *testing.T) {
		args := []string{
			cmdHash,
			"argon2",
			flagPassword, hashPassword,
			flagTest, testArgon2,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash argon2 command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Log(out)
	})
}
