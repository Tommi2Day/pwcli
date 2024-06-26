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
			"hash",
			"--hash-method=ssha",
			"--password", hashPassword,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash ssha command should  not return an error:%s", err)
		assert.Contains(t, out, pwlib.SSHAPrefix, "Output should contain SSHA header")
		t.Logf(out)
	})
	t.Run("TestHashSSHAMatch", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=ssha",
			"--password", hashPassword,
			"--test", testSSHA,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash ssha command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Logf(out)
	})
	t.Run("TestHashMD5", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=md5",
			"--username", hashUsername,
			"--password", hashPassword,
			"--prefix", "md5",
			"--test=",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash md5 command should  not return an error:%s", err)
		assert.Contains(t, out, "md5ebcd", "Output should contain MD5 prefix")
		t.Logf(out)
	})
	t.Run("TestHashMD5Match", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=md5",
			"--username", hashUsername,
			"--password", hashPassword,
			"--prefix", "md5",
			"--test", testMD5,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash md5 command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Logf(out)
	})
	t.Run("TestHashScram", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=scram",
			"--username", hashUsername,
			"--password", hashPassword,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash scram command should  not return an error:%s", err)
		assert.Contains(t, out, "SCRAM-SHA-256$4096:", "Output should contain SCRAM-SHA-256 header")
		t.Logf(out)
	})
	t.Run("TestHashBcryptCompare", func(t *testing.T) {
		var hash string
		hash, err = doBcrypt(hashPassword)
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(hashPassword))
		require.NoErrorf(t, err, "bcrypt compare should  not return an error:%s", err)
	})
	t.Run("TestHashBcrypt", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=bcrypt",
			"--password", hashPassword,
			"--test=",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash bcrypt command should  not return an error:%s", err)
		assert.Contains(t, out, "$2a$", "Output should contain bcrypt header")
		t.Logf(out)
	})
	t.Run("TestHashBcryptMatch", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=bcrypt",
			"--password", hashPassword,
			"--test", testBcrypt,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash bcrypt command should  not return an error:%s", err)
		assert.Contains(t, out, "OK, test input matches", "Output should contain OK message")
		t.Logf(out)
	})
}
