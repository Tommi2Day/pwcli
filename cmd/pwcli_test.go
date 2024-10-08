package cmd

import (
	"os"
	"path"
	"testing"

	"github.com/tommi2day/pwcli/test"

	"github.com/tommi2day/gomodules/common"

	"github.com/tommi2day/gomodules/pwlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
const wrong = "xxx"

// nolint gosec
const totpSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestCLI(t *testing.T) {
	var err error
	var out = ""
	test.InitTestDirs()
	_ = os.RemoveAll(test.TestData)
	_ = os.Mkdir(test.TestData, 0700)
	app := "test_pwcli"
	configFile := path.Join(test.TestData, app+".yaml")
	pc := pwlib.NewConfig(app, test.TestData, test.TestData, app, typeGO)
	err = os.Chdir(test.TestDir)
	require.NoErrorf(t, err, "ChDir failed")
	filename := pc.PlainTextFile
	_ = os.Remove(filename)
	//nolint gosec
	err = common.WriteStringToFile(filename, plain)
	require.NoErrorf(t, err, "Create testdata failed")
	t.Run("CMD GenPass", func(t *testing.T) {
		args := []string{
			"gen",
			"--profile", "10 1 1 1 0 1",
			"--special_chars", "#!",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error:%s", err)
		t.Log(out)
	})
	t.Run("CMD CheckPass default", func(t *testing.T) {
		args := []string{
			"check",
			"--info",
			"--unit-test",
			"NEML2xqZcC",
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Log(out)
	})
	t.Run("CMD CheckCustom OK", func(t *testing.T) {
		args := []string{
			"check",
			"--profile", "4 1 1 0 0 1",
			"--info",
			"--unit-test",
			"qZcC",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Log(out)
	})
	t.Run("CMD CheckPass failure", func(t *testing.T) {
		args := []string{
			"check",
			"--profile", "12 1 1 1 1 1",
			"--special_chars", "#!",
			"--info",
			"--unit-test",
			"NEML2xqZcC",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "Check command should return an error")
		assert.Contains(t, err.Error(), "matches NOT the given profile", "Output should confirm Nomatch")
		t.Log(out)
	})
	t.Run("CMD save config", func(t *testing.T) {
		args := []string{
			"config",
			"save",
			"--config", configFile,
			"--app", app,
			"--datadir", test.TestData,
			"--keydir", test.TestData,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Gen command should not return an error:%s", err)
		assert.Contains(t, out, "config saved to", "Output should confirm saving")
		t.Log(out)
	})
	t.Run("CMD Generate Keypair", func(t *testing.T) {
		args := []string{
			"genkey",
			"--keypass", kp,
			"--method", typeGO,
			"--config", configFile,
			"--app", app,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Generate command should not return an error:%s", err)
		assert.FileExistsf(t, pc.PrivateKeyFile, "Private key file not found")
		assert.FileExistsf(t, pc.PubKeyFile, "Public key file not found")
		assert.Contains(t, out, "New key pair generated as", "Output should confirm key generation")
		t.Log(out)
	})
	t.Run("CMD Encrypt go", func(t *testing.T) {
		args := []string{
			"encrypt",
			"--keypass", kp,
			"--config", configFile,
			"--method", typeGO,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, pc.CryptedFile, "Crypted file '%s' not found", pc.CryptedFile)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Log(out)
	})

	t.Run("CMD Encrypt Openssl", func(t *testing.T) {
		args := []string{
			"encrypt",
			"--keypass", kp,
			"--config", configFile,
			"--method", typeOpenSSL,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		expected := path.Join(test.TestData, app+".pw")
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, expected, "Crypted file '%s' not found", pc.CryptedFile)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Log(out)
	})
	t.Run("CMD decrypt", func(t *testing.T) {
		plaintext := test.TestData + "/plain.txt"
		args := []string{
			"decrypt",
			"--keypass", kp,
			"--method", typeGO,
			"--crypted", pc.CryptedFile,
			"--plaintext", plaintext,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "decrypt command should not return an error:%s", err)
		assert.Contains(t, out, "successfully created", "Output should confirm decryption")
		assert.FileExists(t, plaintext, "Plaintext file %s not found", plaintext)
		c1 := ""
		c2 := ""
		c1, err = common.ReadFileToString(plaintext)
		c2, err = common.ReadFileToString(pc.PlainTextFile)
		assert.Equal(t, c1, c2, "decoded file %s not equal to plaintext file %s", plaintext, pc.PlainTextFile)
		t.Log(out)
	})
	t.Run("CMD list", func(t *testing.T) {
		args := []string{
			"list",
			"--keypass", kp,
			"--config", configFile,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "list command should not return an error:%s", err)
		assert.Contains(t, out, "List returned 10 lines", "Output should lines of plainfile")
		t.Log(out)
	})
	t.Run("CMD get", func(t *testing.T) {
		args := []string{
			"get",
			"--keypass", kp,
			"--config", configFile,
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
	t.Run("CMD get nomatch", func(t *testing.T) {
		args := []string{
			"get",
			"--keypass", kp,
			"--config", configFile,
			"--info",
			"--unit-test",
			"--system", "test",
			"--user", wrong,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "get command should  return an error")
		assert.NotContains(t, out, "Found matching entry", "Output should not confirm match")
		t.Log(out)
	})

	t.Run("CMD TOTP", func(t *testing.T) {
		t.Run("CMD TOTP no secret", func(t *testing.T) {
			_ = os.Unsetenv("TOTP_SECRET")
			out = ""
			args := []string{
				"totp",
				"--info",
				"--unit-test",
			}
			out, err = common.CmdRun(RootCmd, args)
			require.Errorf(t, err, "totp command should return an error")
		})
		t.Run("CMD TOTP Env", func(t *testing.T) {
			_ = os.Setenv("TOTP_SECRET", totpSecret)
			out = ""
			args := []string{
				"totp",
				"--info",
				"--unit-test",
			}
			out, err = common.CmdRun(RootCmd, args)
			require.NoErrorf(t, err, "totp command should  not return an error:%s", err)
			assert.Contains(t, out, "TOTP returned", "Output should confirm success")
			t.Log(out)
		})
		t.Run("CMD TOTP wrong secret", func(t *testing.T) {
			args := []string{
				"totp",
				"--secret", wrong,
				"--info",
				"--unit-test",
			}
			out, err = common.CmdRun(RootCmd, args)
			require.Errorf(t, err, "totp command should return an error")
		})
		t.Run("CMD TOTP with secret", func(t *testing.T) {
			args := []string{
				"totp",
				"--secret", totpSecret,
				"--info",
				"--unit-test",
			}
			out, err = common.CmdRun(RootCmd, args)
			require.NoErrorf(t, err, "totp command should  not return an error:%s", err)
			assert.Contains(t, out, "TOTP returned", "Output should confirm success")
			t.Log(out)
		})
	})
}
