package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
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
	test.InitTestDirs()

	kmsapp := "test_kms_file"
	testdata := test.TestData
	kmspc := pwlib.NewConfig(kmsapp, testdata, testdata, kmsapp, typeKMS)
	_ = os.Chdir(testdata)

	err = os.Chdir(test.TestDir)
	require.NoErrorf(t, err, "ChDir failed")
	filename := kmspc.PlainTextFile
	_ = os.Remove(filename)
	//nolint gosec
	err = common.WriteStringToFile(filename, plain)
	require.NoErrorf(t, err, "Create testdata failed")

	var kmsClient *kms.Client
	kmsContainer, err := prepareKmsContainer()
	defer common.DestroyDockerContainer(kmsContainer)
	require.NoErrorf(t, err, "KMS Server not available")
	require.NotNil(t, kmsContainer, "Prepare failed")
	if err != nil || kmsContainer == nil {
		t.Fatal("KMS server not available")
	}

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
	err = os.Chdir(testdata)
	if err != nil {
		t.Fatal("Change directory failed")
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
			"--plaintext", filename,
			"--kms_keyid", keyID,
			"--method", typeKMS,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Encrypt command should not return an error:%s", err)
		assert.FileExistsf(t, kmspc.CryptedFile, "Crypted file '%s' not found", kmspc.CryptedFile)
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

	t.Run("CMD KMS subcommands", func(t *testing.T) {
		testKMSSubcommands(t)
	})

	// Test KMS Alias Management Commands (reusing the same container)
	t.Run("CMD KMS Alias Management", func(t *testing.T) {
		testKMSAliasManagement(t, kmsClient)
	})
	common.DestroyDockerContainer(kmsContainer)
}

func testKMSSubcommands(t *testing.T) {
	t.Helper()
	var err error
	var out string
	var symKeyID string
	var rsaKeyID string

	t.Run("Generate Symmetric Key", func(t *testing.T) {
		args := []string{
			"kms",
			"generate",
			"--description", "test symmetric key",
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoError(t, err)
		assert.Contains(t, out, "KeyID:")
		re := regexp.MustCompile(`KeyID:\s*([a-f0-9\-]+)`)
		matches := re.FindStringSubmatch(out)
		assert.NotEmpty(t, matches, "Failed to find KeyID in output")
		if len(matches) > 1 {
			symKeyID = matches[1]
			assert.NotEmpty(t, symKeyID, "Extracted KeyID should not be empty")
		}
		t.Log(out)
	})

	t.Run("Generate RSA Key", func(t *testing.T) {
		args := []string{
			"kms",
			"generate",
			"--type", pwlib.KeyTypeRSA,
			"--description", "test rsa key",
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		assert.NoError(t, err)
		assert.Contains(t, out, "KeyID:")
		re := regexp.MustCompile(`KeyID:\s*([a-f0-9\-]+)`)
		matches := re.FindStringSubmatch(out)
		assert.NotEmpty(t, matches, "Failed to find KeyID in output")
		if len(matches) > 1 {
			rsaKeyID = matches[1]
			assert.NotEmpty(t, rsaKeyID, "Extracted KeyID should not be empty")
		}
		t.Log(out)
	})

	t.Run("Generate Key with invalid endpoint", func(t *testing.T) {
		args := []string{
			"kms",
			"generate",
			"--description", "test key with invalid endpoint",
			"--kms_endpoint", "http://invalid-endpoint:1234",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Expected error when endpoint is invalid")
		assert.Contains(t, err.Error(), "failed to call service: KMS", "Error message should indicate connection failure")
		t.Log(out)
	})

	t.Run("Describe KMS Key", func(t *testing.T) {
		args := []string{
			"kms",
			"describe",
			"--kms_keyid", symKeyID,
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, symKeyID)
		t.Log(out)
	})

	t.Run("Test Export symetric KMS Key", func(t *testing.T) {
		args := []string{
			"kms",
			"export",
			"--kms_keyid", symKeyID,
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, "Plaintext Key:")
		t.Log(out)
	})

	t.Run("Test Export RSA KMS Key", func(t *testing.T) {
		args := []string{
			"kms",
			"export",
			"--kms_keyid", rsaKeyID,
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		switch {
		case err != nil && strings.Contains(err.Error(), "UnsupportedOperationException"):
			t.Logf("Skipping export test due to simulator limitation: %v", err)
		case err != nil && strings.Contains(err.Error(), "InvalidKeyUsageException"):
			// Expected: SIGN_VERIFY keys don't support GenerateDataKey
			t.Logf("Expected error for SIGN_VERIFY key: %v", err)
		default:
			require.NoError(t, err)
			assert.Contains(t, out, "RSA Key Export Complete")
			t.Log(out)
		}
	})

	t.Run("Test Get KMS Policy", func(t *testing.T) {
		args := []string{
			"kms",
			"get-policy",
			"--kms_keyid", symKeyID,
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, "Statement")
		t.Log(out)
	})

	t.Run("Test Put KMS Policy", func(t *testing.T) {
		policyFile := "test_policy.json"
		_ = os.WriteFile(policyFile, []byte(out), 0600)
		defer func(name string) {
			_ = os.Remove(name)
		}(policyFile)
		args := []string{
			"kms",
			"put-policy",
			"--kms_keyid", symKeyID,
			"--file", policyFile,
			"--kms_endpoint", kmsAddress,
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, "DONE")
	})

	t.Run("Test Delete KMS Key", func(t *testing.T) {
		args := []string{"kms", "delete", "--kms_keyid", symKeyID, "--days", "7", "--kms_endpoint", kmsAddress}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err)
		assert.Contains(t, out, "DeletionDate:")
		t.Log(out)
	})
}

func testKMSAliasManagement(t *testing.T, kmsClient *kms.Client) {
	t.Helper()
	var err error
	var out string
	var keyID1, keyID2 string

	t.Run("Setup: Create Test Keys for Aliases", func(t *testing.T) {
		keyout1, err := pwlib.GenKMSKey(kmsClient, "", "Test key 1 for alias", map[string]string{"purpose": "alias-test-1"})
		require.NoError(t, err, "Failed to create first test key")
		require.NotNil(t, keyout1)
		keyID1, _ = pwlib.GetKMSKeyIDs(keyout1.KeyMetadata)
		require.NotEmpty(t, keyID1, "KeyID1 should not be empty")

		keyout2, err := pwlib.GenKMSKey(kmsClient, "", "Test key 2 for alias", map[string]string{"purpose": "alias-test-2"})
		require.NoError(t, err, "Failed to create second test key")
		require.NotNil(t, keyout2)
		keyID2, _ = pwlib.GetKMSKeyIDs(keyout2.KeyMetadata)
		require.NotEmpty(t, keyID2, "KeyID2 should not be empty")

		t.Logf("Created test keys: %s and %s", keyID1, keyID2)
	})

	aliasName := "alias/pwcli-test-alias"

	t.Run("Test Create Alias - Success", func(t *testing.T) {
		args := []string{
			"kms", "create-alias",
			"--kms_keyid", keyID1,
			"--alias", aliasName,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err, "Create alias should succeed")
		assert.Contains(t, out, "created successfully", "Output should confirm alias creation")
		assert.Contains(t, out, aliasName, "Output should contain alias name")
		assert.Contains(t, out, keyID1, "Output should contain key ID")
		t.Log(out)
	})

	t.Run("Test Create Alias - Invalid Name (missing alias/ prefix)", func(t *testing.T) {
		args := []string{
			"kms", "create-alias",
			"--kms_keyid", keyID1,
			"--alias", "invalid-alias-name",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Create alias with invalid name should fail")
		assert.Contains(t, err.Error(), "must start with 'alias/'", "Error should mention alias/ requirement")
		t.Log(err)
	})

	t.Run("Test Create Alias - Missing Alias Name", func(t *testing.T) {
		args := []string{
			"kms", "create-alias",
			"--kms_keyid", keyID1,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Create alias without name should fail")
		assert.Contains(t, err.Error(), "alias", "Error should mention alias validation")
		t.Log(err)
	})

	t.Run("Test Create Alias - Missing Key ID", func(t *testing.T) {
		oldKeyIDEnv := os.Getenv("KMS_KEYID")
		_ = os.Unsetenv("KMS_KEYID")
		oldKmsKeyID := kmsKeyID
		kmsKeyID = ""
		viper.Set("kms_keyid", "")
		defer func() {
			if oldKeyIDEnv != "" {
				_ = os.Setenv("KMS_KEYID", oldKeyIDEnv)
			}
			kmsKeyID = oldKmsKeyID
		}()
		args := []string{
			"kms", "create-alias",
			"--alias", "alias/another-test",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Create alias without key ID should fail")
		assert.Contains(t, err.Error(), "kms_keyid", "Error should mention missing key ID")
		t.Log(err)
	})

	t.Run("Test List Aliases - All", func(t *testing.T) {
		args := []string{
			"kms", "list-aliases",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err, "List aliases should succeed")
		assert.Contains(t, out, aliasName, "Output should contain created alias")
		assert.Contains(t, out, keyID1, "Output should contain associated key ID")
		t.Log(out)
	})

	t.Run("Test List Aliases - Filtered by Key ID", func(t *testing.T) {
		args := []string{
			"kms", "list-aliases",
			"--key-id", keyID1,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err, "List filtered aliases should succeed")
		assert.Contains(t, out, aliasName, "Output should contain alias for key1")
		assert.Contains(t, out, keyID1, "Output should contain key1 ID")
		t.Log(out)
	})

	t.Run("Test Update Alias - Success", func(t *testing.T) {
		args := []string{
			"kms", "update-alias",
			"--kms_keyid", keyID2,
			"--alias", aliasName,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err, "Update alias should succeed")
		assert.Contains(t, out, "updated successfully", "Output should confirm alias update")
		assert.Contains(t, out, aliasName, "Output should contain alias name")
		assert.Contains(t, out, keyID2, "Output should contain new key ID")
		t.Log(out)

		argsVerify := []string{
			"kms", "list-aliases",
			"--key-id", keyID2,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		outVerify, errVerify := common.CmdRun(RootCmd, argsVerify)
		require.NoError(t, errVerify, "Verify update should succeed")
		assert.Contains(t, outVerify, aliasName, "Alias should now be associated with keyID2")
		t.Log("Verified: " + outVerify)
	})

	t.Run("Test Update Alias - Invalid Name", func(t *testing.T) {
		args := []string{
			"kms", "update-alias",
			"--kms_keyid", keyID1,
			"--alias", "no-alias-prefix",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Update alias with invalid name should fail")
		assert.Contains(t, err.Error(), "must start with 'alias/'", "Error should mention alias/ requirement")
		t.Log(err)
	})

	t.Run("Test Update Alias - Missing Key ID", func(t *testing.T) {
		oldKeyIDEnv := os.Getenv("KMS_KEYID")
		_ = os.Unsetenv("KMS_KEYID")
		oldKmsKeyID := kmsKeyID
		kmsKeyID = ""
		viper.Set("kms_keyid", "")
		defer func() {
			if oldKeyIDEnv != "" {
				_ = os.Setenv("KMS_KEYID", oldKeyIDEnv)
			}
			kmsKeyID = oldKmsKeyID
		}()
		args := []string{
			"kms", "update-alias",
			"--alias", aliasName,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Update alias without key ID should fail")
		assert.Contains(t, err.Error(), "kms_keyid", "Error should mention missing key ID")
		t.Log(err)
	})

	t.Run("Test Delete Alias - Success", func(t *testing.T) {
		args := []string{
			"kms", "delete-alias",
			"--alias", aliasName,
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoError(t, err, "Delete alias should succeed")
		assert.Contains(t, out, "deleted successfully", "Output should confirm alias deletion")
		assert.Contains(t, out, aliasName, "Output should contain deleted alias name")
		t.Log(out)

		argsVerify := []string{
			"kms", "list-aliases",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		outVerify, errVerify := common.CmdRun(RootCmd, argsVerify)
		require.NoError(t, errVerify, "List aliases should succeed after deletion")
		assert.NotContains(t, outVerify, aliasName, "Deleted alias should not appear in list")
		t.Log("Verified deletion: alias not found in list")
	})

	t.Run("Test Delete Alias - Invalid Name", func(t *testing.T) {
		args := []string{
			"kms", "delete-alias",
			"--alias", "invalid-name",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Delete alias with invalid name should fail")
		assert.Contains(t, err.Error(), "must start with 'alias/'", "Error should mention alias/ requirement")
		t.Log(err)
	})

	t.Run("Test Delete Alias - Missing Alias Name", func(t *testing.T) {
		args := []string{
			"kms", "delete-alias",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Delete alias without name should fail")
		assert.Contains(t, err.Error(), "alias", "Error should mention alias validation")
		t.Log(err)
	})

	t.Run("Test Delete Alias - Non-existent Alias", func(t *testing.T) {
		args := []string{
			"kms", "delete-alias",
			"--alias", "alias/non-existent-alias-xyz",
			"--kms_endpoint", kmsAddress,
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Error(t, err, "Delete non-existent alias should fail")
		assert.Contains(t, err.Error(), "failed to delete alias", "Error should mention deletion failure")
		t.Log(err)
	})

	t.Run("Cleanup: Delete Test Keys", func(t *testing.T) {
		args1 := []string{"kms", "delete", "--kms_keyid", keyID1, "--days", "7", "--kms_endpoint", kmsAddress}
		_, _ = common.CmdRun(RootCmd, args1)

		args2 := []string{"kms", "delete", "--kms_keyid", keyID2, "--days", "7", "--kms_endpoint", kmsAddress}
		_, _ = common.CmdRun(RootCmd, args2)

		t.Log("Test keys scheduled for deletion")
	})
}
