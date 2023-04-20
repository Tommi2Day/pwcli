// Package cmd Commands
package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt plaintext file",
	Long: `Encrypt a plain file given in -p and saved as crypted fin given by -c flag
default for paintext File is <app>.plain and for crypted file is <app.pw>`,
	RunE:         encrypt,
	SilenceUsage: true,
}

func encrypt(cmd *cobra.Command, _ []string) error {
	log.Debug("encrypt called")
	// check for plaintext file option
	pfilename, _ := cmd.Flags().GetString("plaintext")
	if pfilename != "" {
		pc.PlainTextFile = pfilename
	}
	log.Debugf("encrypt plaintext file '%s' with method %s", pc.PlainTextFile, pc.Method)

	// check for crypted file option
	cfilename, _ := cmd.Flags().GetString("crypted")
	if cfilename != "" {
		pc.CryptedFile = cfilename
	}
	log.Debugf("create crypted file '%s'", pc.CryptedFile)

	// check for keypass file option
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate key password '%s'", keypass)
	}

	// do encrypt with default key
	err := pc.EncryptFile()
	if err == nil {
		log.Infof("crypted file '%s' successfully created", pc.CryptedFile)
		fmt.Println("DONE")
	}
	return err
}
func init() {
	RootCmd.AddCommand(encryptCmd)
	// don't have variables populated here
	encryptCmd.PersistentFlags().StringP("plaintext", "t", "", "alternate plaintext file")
	encryptCmd.PersistentFlags().StringP("crypted", "c", "", "alternate crypted file")
	encryptCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
}
