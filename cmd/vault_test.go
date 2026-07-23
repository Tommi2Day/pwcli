package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/pwcli/test"
)

const backtickIDCmd = "`id`"

func TestVault(t *testing.T) {
	var err error
	var out = ""
	test.InitTestDirs()
	if os.Getenv("SKIP_VAULT") != "" {
		t.Skip("Skip Vault Test in CI")
		return
	}
	vaultContainer, pgContainer, err := prepareVaultContainer()
	defer common.DestroyDockerContainer(vaultContainer)
	defer common.DestroyDockerContainer(pgContainer)
	require.NoErrorf(t, err, "Vault Server not available")
	require.NotNil(t, vaultContainer, "Prepare failed")
	if err != nil || vaultContainer == nil {
		t.Fatal("Vault server not available")
	}
	vaulthost, vaultport := common.GetContainerHostAndPort(vaultContainer, "8200/tcp")
	address := fmt.Sprintf("http://%s:%d", vaulthost, vaultport)
	_ = os.Setenv("VAULT_ADDR", address)
	err = os.Setenv("VAULT_TOKEN", rootToken)
	if err != nil {
		t.Fatalf("cannot set vault environment")
	}
	t.Logf("ADDR=%s, Token=%s", address, rootToken)
	t.Run("CMD vault write", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdWrite,
			flagLogicalFalse,
			flagInfo,
			flagUnitTest,
			flagMount, secretMount,
			flagPath, testID,
			flagVaultAddr, address,
			flagVaultToken, rootToken,
			"{\"password\": \"testpass\"}",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Write command should  not return an error: %s", err)
		assert.Contains(t, out, "Vault Write OK", "Output should not confirm success")
		t.Log(out)
	})

	t.Run("CMD vault read", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagLogicalFalse,
			flagInfo,
			flagUnitTest,
			flagMount, secretMount,
			flagPath, testID,
			flagVaultAddr, address,
			flagVaultToken, rootToken,
			entryPassword,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		// Output contains log messages and data. Check for data exactly.
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD vault read json", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagLogicalFalse,
			flagInfo,
			flagUnitTest,
			flagMount, secretMount,
			flagPath, testID,
			"--json",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		assert.True(t, strings.Contains(out, "{"), "Output should be json")
		t.Log(out)
	})
	viper.Reset()
	jsonOut = false
	t.Run("CMD vault read export", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagLogicalFalse,
			flagInfo,
			flagUnitTest,
			flagMount, secretMount,
			flagPath, testID,
			"--json=false",
			"--export",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(out, "testpass"), "Output should contain password")
		assert.True(t, strings.Contains(out, "export PASSWORD='testpass'"), "Output should be export format")
		t.Log(out)
	})
	viper.Reset()
	t.Run("CMD vault list", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdList,
			flagInfo,
			flagUnitTest,
			flagMount, secretMount,
			flagPath, "/",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		expected := "Vault List returned 3 entries"
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Containsf(t, out, expected, "Output should ccontain %s", expected)
	})
	t.Run("CMD vault list demo", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdList,
			flagInfo,
			flagUnitTest,
			flagMount, secretMount,
			flagPath, "demo",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		expected := "Vault List returned 2 entries"
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Containsf(t, out, expected, "Output should ccontain %s", expected)
	})
	t.Run("CMD vault list empty", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdList,
			flagInfo,
			flagUnitTest,
			flagMount, "",
			flagPath, "dummy",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		t.Log(out)
		expected := "Vault List returned 0 entries"
		require.NoErrorf(t, err, "list command should  not return an error:%s", err)
		assert.Containsf(t, out, expected, "Output should ccontain %s", expected)
	})
	viper.Reset()
	t.Run("CMD GetPassword Vault", func(t *testing.T) {
		args := []string{
			cmdGet,
			flagMethod, typeVault,
			flagDebug,
			flagUnitTest,
			flagConfig, test.TestData + "/test_pwcli.yaml",
			flagPath, "secret/data/test",
			"--entry", entryPassword,
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should  not return an error:%s", err)
		assert.Contains(t, out, "Found matching entry", "Output should confirm success")
		t.Log(out)
	})
	t.Run("test removing global options", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdList,
			"--help",
			flagInfo,
			flagUnitTest,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "help command should  not return an error: %s", err)
		assert.Contains(t, out, flagPath, "Output should contain path flag")
		assert.NotContains(t, out, flagDatadir, "Output should not contain datadir flag")
		t.Log(out)
	})

	t.Run("CMD vault read database role demo-ro", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagInfo,
			flagUnitTest,
			flagLogical,
			flagPath, "database/creds/demo-ro",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(strings.ToLower(out), "username"), "Output should contain username")
		assert.True(t, strings.Contains(strings.ToLower(out), entryPassword), "Output should contain password")
		t.Log(out)
	})

	t.Run("CMD vault read database role demo-rw", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagInfo,
			flagUnitTest,
			flagLogical,
			flagPath, "database/creds/demo-rw",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "get command should not return an error:%s", err)
		assert.Contains(t, out, "Vault Data successfully processed", "Output should confirm success")
		assert.True(t, strings.Contains(strings.ToLower(out), "username"), "Output should contain username")
		assert.True(t, strings.Contains(strings.ToLower(out), entryPassword), "Output should contain password")
		t.Log(out)
	})

	t.Run("Connect to Database with Vault Credentials", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagInfo,
			flagUnitTest,
			"--json",
			"--export=false",
			flagLogical,
			flagPath, "database/creds/demo-ro",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "vault read should not return an error: %s", err)
		connectVaultDBCredentials(t, out, pgContainer)
	})

	t.Run("Connect to Database with Vault Export Credentials", func(t *testing.T) {
		args := []string{
			typeVault,
			cmdRead,
			flagInfo,
			flagUnitTest,
			"--export",
			"--json=false",
			flagLogical,
			flagPath, "database/creds/demo-ro",
			flagVaultAddr, address,
			flagVaultToken, rootToken,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "vault read should not return an error: %s", err)
		connectVaultDBExportCredentials(t, out, pgContainer)
	})
}

func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	old := os.Stdout
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	_ = r.Close()
	require.NoError(t, err)
	return string(out)
}

func TestShellEscapeSingleQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no special chars", "simplepass", "simplepass"},
		{"dollar sign", "$HOME", "$HOME"},
		{"backtick", backtickIDCmd, backtickIDCmd},
		{"exclamation", "pass!word", "pass!word"},
		{"single quote", "it's", "it'\\''s"},
		{"double quote", `say "hi"`, `say "hi"`},
		{"backslash", `back\slash`, `back\slash`},
		{"combined", `p@$$'w0rd!`, `p@$$'\''w0rd!`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellEscapeSingleQuote(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintExportOutputShellSafe(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		contains []string
	}{
		{
			name:     "dollar sign not expanded",
			data:     map[string]interface{}{entryPassword: "$HOME"},
			contains: []string{"export PASSWORD='$HOME'"},
		},
		{
			name:     "backtick not executed",
			data:     map[string]interface{}{entryPassword: backtickIDCmd},
			contains: []string{"export PASSWORD='" + backtickIDCmd + "'"},
		},
		{
			name:     "single quote escaped",
			data:     map[string]interface{}{entryPassword: "it's"},
			contains: []string{"export PASSWORD='it'\\''s'"},
		},
		{
			name:     "exclamation mark safe",
			data:     map[string]interface{}{entryPassword: "pass!word"},
			contains: []string{"export PASSWORD='pass!word'"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(t, func() {
				printExportOutput(tt.data)
			})
			for _, expected := range tt.contains {
				assert.Contains(t, out, expected)
			}
		})
	}
}

func TestSanitizeExportKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple key", entryPassword, "PASSWORD"},
		{"already upper", "USER", "USER"},
		{"space", "my key", "MY_KEY"},
		{"semicolon injection", "foo; rm -rf /", "FOO__RM__RF__"},
		{"backtick injection", "foo`id`", "FOO_ID_"},
		{"dollar injection", "foo$(id)", "FOO__ID_"},
		{"newline injection", "foo\nrm -rf /", "FOO_RM__RF__"},
		{"equals sign", "foo=bar", "FOO_BAR"},
		{"leading digit", "1foo", "_1FOO"},
		{"empty key", "", "_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeExportKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintExportOutputKeyInjectionSafe(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		contains []string
		excludes []string
	}{
		{
			name:     "semicolon in key cannot terminate export statement",
			data:     map[string]interface{}{"foo; touch pwned #": "value"},
			contains: []string{"export FOO__TOUCH_PWNED__='value'"},
			excludes: []string{"export FOO; touch pwned #"},
		},
		{
			name:     "newline in key cannot inject a new command",
			data:     map[string]interface{}{"foo\nrm -rf /": "value"},
			contains: []string{"export FOO_RM__RF__='value'"},
		},
		{
			name:     "backtick in key not executed",
			data:     map[string]interface{}{"foo`id`": "value"},
			contains: []string{"export FOO_ID_='value'"},
			excludes: []string{backtickIDCmd},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(t, func() {
				printExportOutput(tt.data)
			})
			for _, expected := range tt.contains {
				assert.Contains(t, out, expected)
			}
			for _, unexpected := range tt.excludes {
				assert.NotContains(t, out, unexpected)
			}
		})
	}
}

func connectVaultDBCredentials(t *testing.T, out string, pgContainer *dockertest.Resource) {
	t.Helper()
	// find the JSON part in the output
	jsonStart := strings.LastIndex(out, "{")
	require.GreaterOrEqual(t, jsonStart, 0, "Output should contain JSON")
	jsonStr := out[jsonStart:]
	// strip trailing log messages if any
	if jsonEnd := strings.LastIndex(jsonStr, "}"); jsonEnd > 0 {
		jsonStr = jsonStr[:jsonEnd+1]
	}

	var rawData map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &rawData), "Failed to unmarshal Vault JSON")

	data := rawData
	if nested, ok := rawData["data"].(map[string]any); ok {
		data = nested
	}
	dbUser, ok1 := data["username"].(string)
	dbPass, ok2 := data[entryPassword].(string)
	require.True(t, ok1 && ok2, "Username or password not found in Vault response")

	pgHost, pgPort := common.GetContainerHostAndPort(pgContainer, "5432/tcp")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/demo?sslmode=disable", dbUser, dbPass, pgHost, pgPort)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to open database connection")
	defer func() { _ = db.Close() }()

	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one), "Failed to execute query with Vault credentials")
	assert.Equal(t, 1, one)
	t.Logf("Successfully connected to database with user: %s", dbUser)
}

func connectVaultDBExportCredentials(t *testing.T, out string, pgContainer *dockertest.Resource) {
	t.Helper()
	var dbUser, dbPass string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export USERNAME=") {
			dbUser = strings.Trim(strings.TrimPrefix(line, "export USERNAME="), "'")
		}
		if strings.HasPrefix(line, "export PASSWORD=") {
			dbPass = strings.Trim(strings.TrimPrefix(line, "export PASSWORD="), "'")
		}
	}
	require.NotEmpty(t, dbUser, "Username not found in Vault export response")
	require.NotEmpty(t, dbPass, "Password not found in Vault export response")

	pgHost, pgPort := common.GetContainerHostAndPort(pgContainer, "5432/tcp")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/demo?sslmode=disable", dbUser, dbPass, pgHost, pgPort)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to open database connection")
	defer func() { _ = db.Close() }()

	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one), "Failed to execute query with Vault export credentials")
	assert.Equal(t, 1, one)
	t.Logf("Successfully connected to database with exported user: %s", dbUser)
}
