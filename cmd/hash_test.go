package cmd

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/tommi2day/gomodules/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hashPassword = "testHashPassword"
const hashUsername = "testHashUsername"

func TestHash(t *testing.T) {
	var out string
	var err error
	t.Run("TestHashMD5", func(t *testing.T) {
		args := []string{
			"hash",
			"--hash-method=md5",
			"--username", hashUsername,
			"--password", hashPassword,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash md5 command should  not return an error:%s", err)
		assert.Contains(t, out, "md5ebcd", "Output should contain md5 prefix")
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
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "hash bcrypt command should  not return an error:%s", err)
		assert.Contains(t, out, "$2a$", "Output should contain bcrypt header")
		t.Logf(out)
	})
}
