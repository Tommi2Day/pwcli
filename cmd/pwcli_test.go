package cmd

import (
	"fmt"
	"os"
	"path"

	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
	"github.com/tommi2day/pwcli/test"
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
const testProfile = `
myprofile:
  # Length Upper Lower Digits Specials FirstIsChar
  profile:
    length: 16
    upper: 1
    lower: 1
    digits: 1
    specials: 0
    first_is_char: true
  special_chars: "!#=@&()"
`

// nolint gosec
const totpSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestCLI(t *testing.T) {
	viper.Reset()
	var err error
	var out = ""
	const testapp = "test_pwcli"
	test.InitTestDirs()
	_ = os.Mkdir(test.TestData, 0700)
	configFile := path.Join(test.TestData, testapp+".yaml")
	err = os.Chdir(test.TestDir)
	require.NoErrorf(t, err, "ChDir failed")
	pc = pwlib.NewConfig(testapp, test.TestData, test.TestData, app, typeGO)
	filename := pc.PlainTextFile
	_ = os.Remove(filename)
	//nolint gosec
	err = common.WriteStringToFile(filename, plain)
	require.NoErrorf(t, err, "Create testdata failed")
	genpassConfig := test.TestData + "/genpass_profile.yeml"
	err = common.WriteStringToFile(genpassConfig, testProfile)
	require.NoErrorf(t, err, "Create testdata failed")
	t.Run("CMD save config to default", func(t *testing.T) {
		const testConfig = "savetest"
		fn := testConfig + ".yaml"
		_ = os.Remove(fn)
		args := []string{
			configKey,
			"save",
			flagApp, testConfig,
			flagDebug,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Save command should not return an error:%s", err)
		assert.Contains(t, out, "config saved to", "Output should confirm saving")
		assert.FileExists(t, fn, "expected config file not found")
		_ = os.Remove(fn)
		t.Log(out)
	})

	t.Run("CMD save config", func(t *testing.T) {
		_ = os.Remove(configFile)
		args := []string{
			configKey,
			"save",
			"--filename", configFile,
			flagApp, testapp,
			flagMethod, typeGO,
			flagDatadir, test.TestData,
			flagKeydir, test.TestData,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Save command should not return an error:%s", err)
		assert.Contains(t, out, "config saved to", "Output should confirm saving")
		t.Log(out)
	})

	t.Run("CMD Get config", func(t *testing.T) {
		args := []string{
			configKey,
			cmdGet,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
			"datadir",
		}
		out, err = common.CmdRun(RootCmd, args)
		expected := fmt.Sprintf("config value for key %s is %s", "datadir", test.TestData)
		require.NoErrorf(t, err, "Get command should not return an error:%s", err)
		assert.Contains(t, out, expected, "Output should contain datadir setting")
		t.Log(out)
	})

	t.Run("CMD print config", func(t *testing.T) {
		args := []string{
			configKey,
			"print",
			flagConfig, configFile,
			flagDebug,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Get command should not return an error:%s", err)
		assert.Contains(t, out, "{", "Output should contain json config")
		t.Log(out)
	})

	t.Run("CMD Generate Keypair", func(t *testing.T) {
		args := []string{
			cmdGenkey,
			flagKeypass, kp,
			flagMethod, typeGO,
			flagConfig, configFile,
			flagApp, testapp,
			flagInfo,
			flagUnitTest,
			flagType, pwlib.KeyTypeRSA,
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
			cmdEncrypt,
			flagApp, testapp,
			flagKeypass, kp,
			flagPlaintext, filename,
			flagMethod, typeGO,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, pc.CryptedFile, "Crypted file '%s' not found", pc.CryptedFile)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Log(out)
	})

	t.Run("CMD Encrypt Openssl", func(t *testing.T) {
		args := []string{
			cmdEncrypt,
			flagKeypass, kp,
			flagPlaintext, filename,
			"--crypted", path.Join(test.TestData, testapp+".pw"),
			flagMethod, typeOpenSSL,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}

		out, err = common.CmdRun(RootCmd, args)
		expected := path.Join(test.TestData, app+".pw")
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, expected, "Crypted file '%s' not found", expected)
		assert.Contains(t, out, "successfully created", "Output should confirm encryption")
		t.Log(out)
	})
	t.Run("CMD decrypt", func(t *testing.T) {
		plaintext := test.TestData + "/plain.txt"
		args := []string{
			"decrypt",
			flagKeypass, kp,
			flagMethod, typeGO,
			flagConfig, configFile,
			flagPlaintext, plaintext,
			flagInfo,
			flagUnitTest,
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
			cmdList,
			flagKeypass, kp,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "list command should not return an error:%s", err)
		assert.Contains(t, out, "List returned 10 lines", "Output should lines of plainfile")
		t.Log(out)
	})
	t.Run("CMD get listmode", func(t *testing.T) {
		args := []string{
			cmdGet,
			"--list",
			flagKeypass, kp,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "list command should not return an error:%s", err)
		assert.Contains(t, out, "List returned 10 lines", "Output should lines of plainfile")
		t.Log(out)
	})
	t.Run("CMD get", func(t *testing.T) {
		args := []string{
			cmdGet,
			"--list=false",
			flagKeypass, kp,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
			flagSystem, testID,
			flagUser, "testuser",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm match")
		assert.Contains(t, out, "'testpass'", "Output should return correct match")
		t.Log(out)
	})
	t.Run("CMD get nomatch", func(t *testing.T) {
		args := []string{
			cmdGet,
			flagKeypass, kp,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
			flagSystem, testID,
			flagUser, wrong,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "get command should  return an error")
		assert.NotContains(t, out, "Found matching entry", "Output should not confirm match")
		t.Log(out)
	})
	t.Run("CMD get sensitive", func(t *testing.T) {
		args := []string{
			cmdGet,
			flagKeypass, kp,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
			flagSystem, testID,
			flagUser, "Testuser",
			"--case-sensitive",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "get command should  return an error")
		assert.NotContains(t, out, "Found matching entry", "Output should not confirm match")
		t.Log(out)
		t.Log(err)
	})

	t.Run("CMD TOTP", func(t *testing.T) {
		t.Run("CMD TOTP no secret", func(t *testing.T) {
			_ = os.Unsetenv("TOTP_SECRET")
			out = ""
			args := []string{
				cmdTotp,
				flagInfo,
				flagUnitTest,
			}
			out, err = common.CmdRun(RootCmd, args)
			require.Errorf(t, err, "totp command should return an error")
		})
		t.Run("CMD TOTP Env", func(t *testing.T) {
			_ = os.Setenv("TOTP_SECRET", totpSecret)
			out = ""
			args := []string{
				cmdTotp,
				flagInfo,
				flagUnitTest,
			}
			out, err = common.CmdRun(RootCmd, args)
			require.NoErrorf(t, err, "totp command should  not return an error:%s", err)
			assert.Contains(t, out, "TOTP returned", "Output should confirm success")
			t.Log(out)
		})
		t.Run("CMD TOTP wrong secret", func(t *testing.T) {
			args := []string{
				cmdTotp,
				"--secret", wrong,
				flagInfo,
				flagUnitTest,
			}
			out, err = common.CmdRun(RootCmd, args)
			require.Errorf(t, err, "totp command should return an error")
		})
		t.Run("CMD TOTP with secret", func(t *testing.T) {
			args := []string{
				cmdTotp,
				"--secret", totpSecret,
				flagInfo,
				flagUnitTest,
			}
			out, err = common.CmdRun(RootCmd, args)
			require.NoErrorf(t, err, "totp command should  not return an error:%s", err)
			assert.Contains(t, out, "TOTP returned", "Output should confirm success")
			t.Log(out)
		})
	})
	// Modify config file with profile
	cf, err := common.ReadFileToString(configFile)
	if err != nil {
		t.Errorf("Error reading config file: %s", err)
	}
	cf += fmt.Sprintf("\n\npassword_profiles: %s\n", genpassConfig)
	err = common.WriteStringToFile(configFile, cf)
	if err != nil {
		t.Errorf("Error modifiying config file: %s", err)
	}
	t.Run("CMD GenPass", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagProfile, "10 1 1 1 0 1",
			"--special_chars", "#!",
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error: %s", err)
		t.Log(out)
	})

	t.Run("CMD CheckPass nopassword", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.Error(t, err, "Check command should  return an error")
		if err != nil {
			assert.Contains(t, err.Error(), "requires password", "error message should contain requires")
		}
	})
	t.Run("CMD CheckPass default", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagInfo,
			flagUnitTest,
			"Idt3P#v2tgEfW0Cx",
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Log(out)
	})
	t.Run("CMD CheckCustom OK", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagProfile, "4 1 1 0 0 1",
			flagInfo,
			flagUnitTest,
			"qZcC",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Log(out)
	})
	t.Run("CMD CheckPass failure", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagProfile, "12 1 1 1 1 1",
			"--special_chars", "#!",
			flagInfo,
			flagUnitTest,
			"NEML2xqZcC",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "Check command should return an error")
		assert.Contains(t, err.Error(), "matches NOT the given profile", "Output should confirm Nomatch")
		t.Log(out)
	})

	t.Run("CMD GenPass default profile", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagProfile, "",
			flagProfileset, "",
			flagConfig, configFile,
			flagDebug,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error: %s", err)
		t.Log(out)
	})
	t.Run("CMD GenPass easy", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagConfig, configFile,
			flagProfile, "",
			flagProfileset, "easy",
			flagDebug,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error: %s", err)
		t.Log(out)
	})
	t.Run("CMD GenPass myprofile", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagConfig, "",
			"--password_profiles", genpassConfig,
			flagProfile, "",
			flagProfileset, "myprofile",
			flagDebug,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error: %s", err)
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD GenPass missing profileset", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagProfileset, "NotExistent",
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.Errorf(t, err, "Gen command should return an error")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD GenPass invalid profileset", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagProfileset, "invalid",
			flagConfig, configFile,
			flagDebug,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.Errorf(t, err, "Gen command should return an error")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD Check local Profileset OK", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagProfile, "",
			flagProfileset, "myprofile",
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
			"LP9w81EiS!usR##R",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD Check Default Profileset OK", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagProfile, "",
			flagProfileset, "",
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
			"LP9w81EiS!usR##R",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Check command should not return an error:%s", err)
		assert.Contains(t, out, "matches the given profile", "Output should confirm match")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD CheckPass profileset fail", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagProfile, "",
			flagConfig, configFile,
			flagProfileset, "myprofile",
			flagInfo,
			flagUnitTest,
			"NEML2xqZcC",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "Check command should return an error")
		assert.Contains(t, err.Error(), "matches NOT the given profile", "Output should confirm Nomatch")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD GenPass list profiles", func(t *testing.T) {
		args := []string{
			cmdGen,
			flagProfile, "",
			flagProfileset, "",
			"--list_profiles",
			"--password_profiles", genpassConfig,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error: %s", err)
		assert.Contains(t, out, "myprofile", "Output should contain myprofile")
		assert.Contains(t, out, "easy", "Output should contain easy")
		t.Log(out)
	})
	t.Run("CMD check list profiles", func(t *testing.T) {
		args := []string{
			cmdCheck,
			flagProfile, "",
			flagProfileset, "",
			"--list_profiles",
			"--password_profiles", genpassConfig,
			flagConfig, configFile,
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoErrorf(t, err, "Gen command should not return an error: %s", err)
		assert.Contains(t, out, "myprofile", "Output should contain myprofile")
		assert.Contains(t, out, "easy", "Output should contain easy")
		t.Log(out)
	})
}
