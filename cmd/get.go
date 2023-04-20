// Package cmd commands
package cmd

import (
	"fmt"

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
	log.Debug("Get password called")
	system, _ := cmd.Flags().GetString("system")
	if system == "" {
		system, _ = cmd.Flags().GetString("db")
	}
	account, _ := cmd.Flags().GetString("user")
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate password: %s", kp)
	}
	pwlib.SilentCheck = false
	password, err := pc.GetPassword(system, account)
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
	_ = getCmd.MarkFlagRequired("user")
	getCmd.MarkFlagsMutuallyExclusive("system", "db")
	// don't have variables populated here
}
