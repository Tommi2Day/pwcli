package cmd

import (
	"fmt"
	"os"
	"testing"
	"time"

	ldap "github.com/go-ldap/ldap/v3"

	"github.com/go-git/go-git/v5/utils/ioutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/ldaplib"
	"github.com/tommi2day/pwcli/test"
)

const LdapBaseDn = "dc=example,dc=local"
const LdapAdminUser = "cn=admin," + LdapBaseDn
const LdapAdminPassword = "admin"
const LdapConfigPassword = "config"
const LdapTestUserDN = "cn=test,ou=Users," + LdapBaseDn
const LdapTestUserPassword = "test"
const LdapTestUser2DN = "cn=test2,ou=Users," + LdapBaseDn
const LdapTestUser2Password = "test2"
const LdapNewPassword = "newpassword"

var sshkey = `ssh-rsa AAAAB3NzaC1yc2EAABBDAQABAFFBAQDX9uIBEySOR6tASa4RIgUo6TTOi+o3hIWkxJGlfajHQY9f73LONotAPKgDfEdvrvY+0
pW3zXe3pmr4GhQzP2c1EMYmwVdkkQLtn/FHFICLHCyihN2byMe14v4iv1em6XXLqVB7cbxi2XKHHfa50tqgeEJTIRVbFVht9WCdHQ9VUvwnCUda6wDt3E1q+
tAaUOldrfFl3KR4LQThOUOEOtaG1eU2Q/fk1j5qLMH2sDtzYnTp2MgLVAElC7XH9QDWz4+I3uxeYOweUhvBnBx+Ti2ZkzZjchRbkawa4v/woySmWove7nzp
BPYWJ8mBdRecVfcY+/jZDSe2Phgfzgf3cRTvs3tF test@ldap`
var sshkey2 = `ssh-rsa AAAAB3NzaC1yc2EAABBDAQABAFFBAQDX9uIBEySOR6tASa4RIgUo6TTOi+o3hIWkxJGlfajHQY9f73LONotAPKgDfEdvrvY+0
pW3zXe3pmr4GhQzP2c1EMYmwVdkkQLtn/FHFICLHCyihN2byMe14v4iv1em6XXLqVB7cbxi2XKHHfa50tqgeEJTIRVbFVht9WCdHQ9VUvwnCUda6wDt3E1q+
tAaUOldrfFl3KR4LQThOUOEOtaG1eU2Q/fk1j5qLMH2sDtzYnTp2MgLVAElC7XH9QDWz4+I3uxeYOweUhvBnBx+Ti2ZkzZjchRbkawa4v/woySmWove7nzp
BPYWJ8mBdRecVfcY+/jZDSe2Phgfzgf3cRTvs3tF test2@ldap`
var lc *ldaplib.LdapConfigType

func TestLdap(t *testing.T) {
	var err error
	var server string
	var sslport int
	var out string

	test.Testinit(t)
	err = os.Chdir(test.TestDir)
	require.NoErrorf(t, err, "ChDir failed")
	ldapAdmin := test.TestData
	sshkeyfile := ldapAdmin + "/id_rsa.pub"
	sshkeyfile2 := ldapAdmin + "/id_rsa2.pub"
	//nolint gosec
	err = os.WriteFile(sshkeyfile, []byte(sshkey), 0644)
	require.NoErrorf(t, err, "Create test id_rsa.pub failed")
	//nolint gosec
	err = os.WriteFile(sshkeyfile2, []byte(sshkey2), 0644)
	require.NoErrorf(t, err, "Create test id_rsa2.pub failed")

	// redirect Stdin for test
	r, w, err := os.Pipe()
	require.NoErrorf(t, err, "Pipe failed")
	inputReader = r

	// prepare or skip container based tests
	if os.Getenv("SKIP_LDAP") != "" {
		t.Skip("Skipping LDAP testing in CI environment")
	}
	ldapContainer, err = prepareLdapContainer()
	require.NoErrorf(t, err, "Ldap Server not available")
	require.NotNil(t, ldapContainer, "Prepare failed")
	defer common.DestroyDockerContainer(ldapContainer)
	server, sslport = common.GetContainerHostAndPort(ldapContainer, "1636/tcp")

	base := LdapBaseDn
	lc = ldaplib.NewConfig(server, sslport, true, true, base, ldapTimeout)

	t.Run("Ldap Connect", func(t *testing.T) {
		t.Logf("Connect '%s' using SSL on port %d", LdapTestUserDN, sslport)
		err = lc.Connect(LdapTestUserDN, LdapTestUserPassword)
		require.NoErrorf(t, err, "admin Connect returned error %v", err)
		assert.NotNilf(t, lc.Conn, "Ldap Connect is nil")
		assert.IsType(t, &ldap.Conn{}, lc.Conn, "returned object ist not ldap connection")
		if lc.Conn == nil {
			t.Fatalf("No valid Connection, terminate")
			return
		}
	})
	t.Run("change NonAdmin ssh key", func(t *testing.T) {
		args := []string{
			"ldap",
			"change-sshpubkey",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapTestUserDN,
			"--ldap.bindpassword", LdapTestUserPassword,
			"--sshpubkeyfile", sshkeyfile,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "SUCCESS: ", "Output not as expected")
		_ = ldapPassCmd.Flags().Set("new_password", "")
	})
	t.Run("set SSH Key by Admin", func(t *testing.T) {
		args := []string{
			"ldap",
			"change-sshpubkey",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapAdminUser,
			"--ldap.bindpassword", LdapAdminPassword,
			"--ldap.targetdn", LdapTestUserDN,
			"--sshpubkeyfile", sshkeyfile,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "SUCCESS: ", "Output not as expected")
		_ = ldapSSHCmd.Flags().Set("ldap.targetdn", "")
	})
	t.Run("set SSH Key without class", func(t *testing.T) {
		args := []string{
			"ldap",
			"change-sshpubkey",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapTestUser2DN,
			"--ldap.bindpassword", LdapTestUser2Password,
			"--sshpubkeyfile", sshkeyfile,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.Errorf(t, err, "Command should return error")
		t.Logf(out)
		assert.Containsf(t, out, "ssh key attribute", "Output not as expected")
		_ = ldapPassCmd.Flags().Set("ldap.targetdn", "")
	})
	t.Run("change NonAdmin Ldap password", func(t *testing.T) {
		args := []string{
			"ldap",
			"change-password",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapTestUserDN,
			"--ldap.bindpassword", LdapTestUserPassword,
			"--new-password", LdapNewPassword,
			"--ldap.targetdn", "",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "SUCCESS: ", "Output not as expected")
		_ = ldapPassCmd.Flags().Set("new-password", "")
		_ = ldapPassCmd.Flags().Set("ldap.targetdn", "")
	})
	t.Run("change Ldap password by Admin", func(t *testing.T) {
		args := []string{
			"ldap",
			"change-password",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapAdminUser,
			"--ldap.bindpassword", LdapAdminPassword,
			"--ldap.targetdn", LdapTestUser2DN,
			"--new-password", LdapNewPassword,
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "SUCCESS: ", "Output not as expected")
		_ = ldapPassCmd.Flags().Set("new-password", "")
		_ = ldapPassCmd.Flags().Set("ldap.targetdn", "")
	})
	t.Run("change Ldap password by Admin with prompt", func(t *testing.T) {
		args := []string{
			"ldap",
			"change-password",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapAdminUser,
			"--ldap.bindpassword", LdapAdminPassword,
			"--ldap.targetdn", LdapTestUser2DN,
			// "--new-password", LdapNewPassword,
			"--info",
			"--unit-test",
		}
		_, _ = w.WriteString(fmt.Sprintf("%s\n", LdapNewPassword+"1"))
		time.Sleep(1 * time.Second)
		_, _ = w.WriteString(fmt.Sprintf("%s\n", LdapNewPassword+"1"))
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "SUCCESS: ", "Output not as expected")
		_ = ldapPassCmd.Flags().Set("new-password", "")
		_ = ldapPassCmd.Flags().Set("ldap.targetdn", "")
	})
	t.Run("change and generate password", func(t *testing.T) {
		np := ldapPassCmd.Flags().Lookup("new-password")
		np.Changed = false
		args := []string{
			"ldap",
			"change-password",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.base", LdapBaseDn,
			"--ldap.binddn", LdapAdminUser,
			"--ldap.bindpassword", LdapAdminPassword,
			"--ldap.targetdn", LdapTestUser2DN,
			"--generate",
			"--profile", "6 1 1 1 0 0",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "SUCCESS: ", "Output not as expected")
		assert.Containsf(t, out, "generated Password: ", "Output not as expected")
		_ = ldapPassCmd.Flags().Set("ldap.targetdn", "")
		targetDN = ""
	})
	t.Run("Show Attributes without basedn", func(t *testing.T) {
		ldapBaseDN = ""
		args := []string{
			"ldap",
			"show",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.binddn", LdapAdminUser,
			"--ldap.bindpassword", LdapAdminPassword,
			"--ldap.targetuser", "test2",
			"--info",
			"--unit-test",
		}
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "sn: test2", "Output not as expected")
	})
	t.Run("Show groups without bind pass", func(t *testing.T) {
		ldapBaseDN = ""
		ldapBindPassword = ""
		args := []string{
			"ldap",
			"groups",
			"--ldap.host", server,
			"--ldap.port", fmt.Sprintf("%d", sslport),
			"--ldap.tls", "true",
			"--ldap.insecure", "true",
			"--ldap.binddn", LdapAdminUser,
			// "--ldap.bindpassword", LdapAdminPassword,
			"--ldap.targetuser", "test2",
			"--info",
			"--unit-test",
		}

		// write to Stdin
		_, _ = w.WriteString(fmt.Sprintf("%s\n", LdapAdminPassword))
		out, err = common.CmdRun(RootCmd, args)
		require.NoErrorf(t, err, "Command returned error: %s", err)
		t.Logf(out)
		assert.Containsf(t, out, "Group: cn=ssh,ou=Groups,dc=example,dc=local", "Output not as expected")
	})
	_ = w.Close()
}

// write unit test for promptPassword
func TestPromptPassword(t *testing.T) {
	// redirect Stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	inputReader = r
	defer func() {
		_ = ioutil.WriteNopCloser(w)
		os.Stdin = oldStdin
	}()
	// write to Stdin
	_, _ = w.WriteString("test\n")
	_ = w.Close()

	// test promptPassword
	password, err := promptPassword("TestPromptPassword:")
	require.NoError(t, err, "PromptPassword should not return error")
	assert.Equal(t, "test", password, "PromptPassword should return test")
}
