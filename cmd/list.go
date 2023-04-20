// Package cmd commands
package cmd

import (
	"fmt"

	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "list passwords",
	Long:         `List all available password records`,
	SilenceUsage: true,
	RunE:         listpass,
}

func listpass(cmd *cobra.Command, _ []string) error {
	log.Debug("list called")
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate password: %s", kp)
	}
	pwlib.SilentCheck = false
	lines, err := pc.ListPasswords()
	if err == nil {
		for _, l := range lines {
			fmt.Println(l)
		}
		log.Infof("List returned %d lines", len(lines))
	}
	return err
}

func init() {
	RootCmd.AddCommand(listCmd)
	// don't have variables populated here
	listCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
}
