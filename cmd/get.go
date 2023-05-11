// Package cmd commands
package cmd

import (
	"fmt"
	"os"

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

func getpass(cmd *cobra.Command, _ []string) error {
	var system string
	var account string
	var password string
	var err error
	log.Debugf("Get password called, method %s", method)
	if method == typeVault {
		account, _ = cmd.Flags().GetString("entry")
		system, _ = cmd.Flags().GetString("path")
		log.Debugf("use vault method with path %s and key %s", system, account)
		if account == "" || system == "" {
			err = fmt.Errorf("method vault needs parameter path and entry set")
			return err
		}
		if vaultAddr != "" {
			_ = os.Setenv("VAULT_ADDR", vaultAddr)
		}
		if vaultToken != "" {
			_ = os.Setenv("VAULT_TOKEN", vaultToken)
		}
	} else {
		system, _ = cmd.Flags().GetString("system")
		if system == "" {
			system, _ = cmd.Flags().GetString("db")
		}
		account, _ = cmd.Flags().GetString("user")
		if account == "" {
			err = fmt.Errorf("need parameter user to proceed")
			return err
		}
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
	getCmd.Flags().StringVarP(&vaultAddr, "vault_addr", "A", vaultAddr, "VAULT_ADDR Url")
	getCmd.Flags().StringVarP(&vaultToken, "vault_token", "T", vaultToken, "VAULT_TOKEN")
}
