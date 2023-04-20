package test

import (
	"os"
	"path"
	"testing"

	"github.com/tommi2day/gomodules/pwlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const typeGO = "go"
const typeOpenssl = "openssl"
const plain = `
# Testfile
!default:defuser2:failure
!default:testuser:default
test:testuser:testpass
testdp:testuser:xxx:yyy
!default:defuser2:default
!default:testuser:failure
!default:defuser:default
`
const kp = "pwcli_test"

func TestCLI(t *testing.T) {
	var err error
	var out = ""
	Testinit(t)
	_ = os.RemoveAll(TestData)
	_ = os.Mkdir(TestData, 0700)
	app := "test_pwcli"
	configFile := path.Join(TestData, app+".yaml")
	pc := pwlib.NewConfig(app, TestData, TestData, app, typeGO)
	err = os.Chdir(TestDir)
	require.NoErrorf(t, err, "ChDir failed")
	filename := pc.PlainTextFile
	_ = os.Remove(filename)
	//nolint gosec
	err = os.WriteFile(filename, []byte(plain), 0644)
	require.NoErrorf(t, err, "Create testdata failed")
	t.Run("CMD GenPass", func(t *testing.T) {
		args := []string{
			"gen",
			"--profile", "10 1 1 1 0 1",
			"--special_chars", "#!",
			"--info",
		}
		out, err = cmdTest(args)
		assert.NoErrorf(t, err, "Gen command should not return an error:%s", err)
		t.Logf(out)
	})
	t.Run("CMD CheckPass default", func(t *testing.T) {
		args := []string{
			"check",
			"--info",
			"NEML2xqZcC",
		}
		out, err = cmdTest(args)
		assert.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Logf(out)
	})
	t.Run("CMD CheckCustom OK", func(t *testing.T) {
		args := []string{
			"check",
			"--profile", "4 1 1 0 0 1",
			"--info",
			"qZcC",
		}
		out, err = cmdTest(args)
		require.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Logf(out)
	})
	t.Run("CMD CheckPass failure", func(t *testing.T) {
		args := []string{
			"check",
			"--profile", "12 1 1 1 1 1",
			"--special_chars", "#!",
			"--info",
			"NEML2xqZcC",
		}
		out, err = cmdTest(args)
		require.Errorf(t, err, "Check command should return an error")
		assert.Contains(t, err.Error(), "matches NOT the given profile", "Output should confirm Nomatch")
		t.Logf(out)
	})
	t.Run("CMD save config", func(t *testing.T) {
		args := []string{
			"config",
			"save",
			"--config", configFile,
			"--app", app,
			"--datadir", TestData,
			"--keydir", TestData,
			"--info",
		}
		out, err = cmdTest(args)
		require.NoErrorf(t, err, "Gen command should not return an error:%s", err)
		assert.Contains(t, out, "config saved to", "Output should confirm saving")
		t.Logf(out)
	})
	t.Run("CMD Generate Keypair", func(t *testing.T) {
		args := []string{
			"genkey",
			"--keypass", kp,
			"--config", configFile,
			"--app", app,
			"--info",
		}
		out, err = cmdTest(args)
		require.NoErrorf(t, err, "Generate command should not return an error:%s", err)
		assert.FileExistsf(t, pc.PrivateKeyFile, "Private key file not found")
		assert.FileExistsf(t, pc.PubKeyFile, "Public key file not found")
		assert.Contains(t, out, "New key pair generated as", "Output should confirm key generation")
		t.Logf(out)
	})
	t.Run("CMD Encrypt default", func(t *testing.T) {
		args := []string{
			"encrypt",
			"--keypass", kp,
			"--config", configFile,
			"--info",
		}
		out, err = cmdTest(args)
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, pc.CryptedFile, "Crypted file '%s' not found", pc.CryptedFile)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Logf(out)
	})

	t.Run("CMD Encrypt Openssl", func(t *testing.T) {
		args := []string{
			"encrypt",
			"--keypass", kp,
			"--config", configFile,
			"--method", typeOpenssl,
			"--info",
		}
		out, err = cmdTest(args)
		expected := path.Join(TestData, app+".pw")
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, expected, "Crypted file '%s' not found", pc.CryptedFile)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Logf(out)
	})

	t.Run("CMD list", func(t *testing.T) {
		args := []string{
			"list",
			"--keypass", kp,
			"--config", configFile,
			"--info",
		}
		out, err = cmdTest(args)
		require.NoErrorf(t, err, "list command should not return an error:%s", err)
		assert.Contains(t, out, "List returned 10 lines", "Output should lines of plainfile")
		t.Logf(out)
	})
	t.Run("CMD get", func(t *testing.T) {
		args := []string{
			"get",
			"--keypass", kp,
			"--config", configFile,
			"--info",
			"--system", "test",
			"--user", "testuser",
		}
		out, err = cmdTest(args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm match")
		assert.Contains(t, out, "'testpass'", "Output should return correct match")
		t.Logf(out)
	})
	t.Run("CMD get nomatch", func(t *testing.T) {
		args := []string{
			"get",
			"--keypass", kp,
			"--config", configFile,
			"--info",
			"--system", "test",
			"--user", "xxx",
		}
		out, err = cmdTest(args)
		require.Errorf(t, err, "get command should  return an error")
		assert.NotContains(t, out, "Found matching entry", "Output should not confirm match")
		t.Logf(out)
	})
}
