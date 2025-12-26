package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/tommi2day/gomodules/ldaplib"
	"github.com/tommi2day/pwcli/test"

	"github.com/tommi2day/gomodules/common"

	"github.com/go-ldap/ldap/v3"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const Ldaprepo = "docker.io/cleanstart/openldap"
const LdaprepoTag = "2.6.10"
const LdapcontainerTimeout = 120

var ldapcontainerName string
var ldapContainer *dockertest.Resource

// prepareContainer create an OpenLdap Docker Container
func prepareLdapContainer() (container *dockertest.Resource, err error) {
	if os.Getenv("SKIP_LDAP") != "" {
		err = fmt.Errorf("skipping LDAP Container in CI environment")
		return
	}
	ldapcontainerName = os.Getenv("LDAP_CONTAINER_NAME")
	if ldapcontainerName == "" {
		ldapcontainerName = "pwcli-ldap"
	}

	var pool *dockertest.Pool
	pool, err = common.GetDockerPool()
	if err != nil || pool == nil {
		return
	}
	vendorImagePrefix := os.Getenv("VENDOR_IMAGE_PREFIX")
	repoString := vendorImagePrefix + Ldaprepo

	fmt.Printf("Try to start docker container for %s:%s\n", repoString, LdaprepoTag)
	fmt.Println(path.Join(test.TestDir, "docker", "ldap", "certs") + ":/certs:ro")
	container, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: repoString,
		Tag:        LdaprepoTag,

		Mounts: []string{
			test.TestDir + "/docker/ldap/certs:/certs:ro",
			// test.TestDir + "/docker/ldap/schema:/schema:ro",
			test.TestDir + "/docker/ldap/ldif:/ldif:ro",
			test.TestDir + "/docker/ldap/etc/slapd.conf:/etc/openldap/slapd.conf:ro",
		},

		Hostname: ldapcontainerName,
		Name:     ldapcontainerName,
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		err = fmt.Errorf("error starting ldap docker container: %v", err)
		return
	}

	pool.MaxWait = LdapcontainerTimeout * time.Second
	myhost, myport := common.GetContainerHostAndPort(container, "389/tcp")
	dialURL := fmt.Sprintf("ldap://%s:%d", myhost, myport)
	fmt.Printf("Wait to successfully connect to Ldap with %s (max %ds)...\n", dialURL, LdapcontainerTimeout)
	start := time.Now()
	var l *ldap.Conn
	if err = pool.Retry(func() error {
		l, err = ldap.DialURL(dialURL)
		return err
	}); err != nil {
		fmt.Printf("Could not connect to LDAP Container: %s", err)
		return
	}
	_ = l.Close()
	elapsed := time.Since(start)
	fmt.Printf("LDAP Container is available after %s\n", elapsed.Round(time.Millisecond))
	time.Sleep(10 * time.Second)
	err = applyLdapConfigs(myhost, myport, test.TestDir+"/docker/ldap/ldif")
	if err != nil {
		return
	}
	err = nil
	return
}

func applyLdapConfigs(server string, port int, ldifDir string) (err error) {
	lpc := ldaplib.NewConfig(server, port, false, false, "cn=config", ldapTimeout)
	err = lpc.Connect(LdapConfigUser, LdapConfigPassword)
	if err != nil || lpc.Conn == nil {
		err = fmt.Errorf("LDAP Config Connect failed: %v", err)
		return
	}

	pattern := "*.schema"
	// Apply all files matching *.config
	err = lpc.ApplyLDIFDir(ldifDir, pattern, false)
	if err != nil {
		return
	}

	// Verify by searching for one of the applied schemas/configs if needed
	// For example, checking if a specific schema DN exists
	schemaBase := "cn=schema,cn=config"
	entries, e := lpc.Search(schemaBase, "(cn=*ldapPublicKey)", []string{"dn"}, ldap.ScopeWholeSubtree, ldap.DerefInSearching)
	if e != nil || len(entries) == 0 {
		err = fmt.Errorf("Search for schema ldapPublicKey failed: %v", e)
		return
	}
	fmt.Printf("Schema Verified: %s exists\n", entries[0].DN)
	pattern = "*.config"
	err = lpc.ApplyLDIFDir(ldifDir, pattern, false)
	if err != nil {
		return
	}
	fmt.Println("LDAP Configs applied")

	// apply entries
	la := ldaplib.NewConfig(server, port, false, false, LdapBaseDn, ldapTimeout)
	err = la.Connect(LdapAdminUser, LdapAdminPassword)
	if err != nil || la.Conn == nil {
		err = fmt.Errorf("LDAP Admin Connect failed: %v", err)
		return
	}
	pattern = "*.ldif"
	err = la.ApplyLDIFDir(ldifDir, pattern, false)
	if err != nil {
		return
	}
	fmt.Println("LDAP Entries prepared")
	return
}
