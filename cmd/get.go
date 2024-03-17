// Package cmd commands
package cmd

import (
	"fmt"
	"os"

	"github.com/tommi2day/gomodules/common"

	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:          "get",
	Short:        "Get encrypted password",
	Long:         `Return a password for a an Account on a system/database`,
	RunE:         getpass,
	SilenceUsage: true,
}

func handleVault(cmd *cobra.Command, account *string, system *string) (err error) {
	*account, _ = cmd.Flags().GetString("entry")
	*system, _ = cmd.Flags().GetString("path")
	log.Debugf("use vault method with path %s and key %s", *system, *account)
	if *account == "" || *system == "" {
		err = fmt.Errorf("method vault needs parameter path and entry set")
		return err
	}
	if vaultAddr != "" {
		_ = os.Setenv("VAULT_ADDR", vaultAddr)
	}
	if vaultToken != "" {
		_ = os.Setenv("VAULT_TOKEN", vaultToken)
	}
	return
}

func handleKMS() (err error) {
	if kmsKeyID == "" {
		kmsKeyID = common.GetStringEnv("KMS_KEYID", "")
		log.Debugf("KMS KeyID from environment: '%s'", kmsKeyID)
	}
	if kmsKeyID == "" {
		return fmt.Errorf("need parameter kms_keyid to proceed")
	}
	if kmsEndpoint != "" {
		log.Debugf("use KMS endpoint %s", kmsEndpoint)
		_ = os.Setenv("KMS_ENDPOINT", kmsEndpoint)
	}
	log.Debugf("use KMS method with keyid %s", kmsKeyID)
	pc.KMSKeyID = kmsKeyID
	return nil
}
func getpass(cmd *cobra.Command, _ []string) error {
	var system string
	var account string
	var password string
	var err error
	log.Debugf("Get password called, method %s", method)
	system, _ = cmd.Flags().GetString("system")
	if system == "" {
		system, _ = cmd.Flags().GetString("db")
	}
	account, _ = cmd.Flags().GetString("user")
	switch method {
	case typeVault:
		err = handleVault(cmd, &account, &system)
	case typeKMS:
		err = handleKMS()
	}
	if err != nil {
		return err
	}
	if account == "" {
		err = fmt.Errorf("need parameter user to proceed")
		return err
	}
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate password: %s", kp)
	}
	pwlib.SilentCheck = false
	password, err = pc.GetPassword(system, account)
	if err == nil {
		fmt.Println(password)
		log.Infof("Found matching entry: '%s'", password)
	}
	return err
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("system", "s", "", "name of the system/database")
	getCmd.Flags().StringP("db", "d", "", "name of the system/database")
	getCmd.Flags().StringP("user", "u", "", "account/user name")
	getCmd.Flags().StringP("keypass", "p", "", "password for the private key")
	getCmd.Flags().StringP("path", "P", "", "vault path to the secret, eg /secret/data/... within method vault, use together with path")
	getCmd.Flags().StringP("entry", "E", "", "vault secret entry key within method vault, use together with path")
	getCmd.Flags().StringVar(&vaultAddr, "vault_addr", vaultAddr, "VAULT_ADDR Url")
	getCmd.Flags().StringVar(&vaultToken, "vault_token", vaultToken, "VAULT_TOKEN")
	getCmd.Flags().StringVar(&kmsKeyID, "kms_keyid", kmsKeyID, "KMS KeyID")
	getCmd.Flags().StringVar(&kmsEndpoint, "kms_endpoint", kmsEndpoint, "KMS Endpoint Url")
}
