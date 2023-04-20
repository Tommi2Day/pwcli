// Package cmd Commands
package cmd

import (
	"fmt"

	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "genkey",
	Aliases: []string{"genrsa"},
	Short:   "Generate a new RSA Keypair",
	Long: `Generates a new pair of rsa keys
optionally you may assign an idividal key password using -p flag
`,
	RunE:         genkey,
	SilenceUsage: true,
}

func genkey(cmd *cobra.Command, _ []string) error {
	log.Debug("generate keys called")
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate key password '%s'", kp)
	}
	_, _, err := pwlib.GenRsaKey(pc.PubKeyFile, pc.PrivateKeyFile, pc.KeyPass)
	if err == nil {
		log.Infof("New key pair generated as %s and %s", pc.PubKeyFile, pc.PrivateKeyFile)
		fmt.Println("DONE")
	}
	return err
}

func init() {
	RootCmd.AddCommand(generateCmd)
	// don't have variables populated here
	generateCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
}
