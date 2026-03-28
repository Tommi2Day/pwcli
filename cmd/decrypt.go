// Package cmd Commands
package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/common"
)

// encryptCmd represents the encrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt crypted file",
	Long: `Decrypt a crypted file given in -c and saved as plaintext file given by -p flag using given method.
default for plaintext File is <app>.plain and for crypted file is <app.pw>`,
	RunE:         decrypt,
	SilenceUsage: true,
}

func decrypt(cmd *cobra.Command, _ []string) error {
	log.Debug("decrypt called")
	// check for crypted file option
	cfilename, _ := cmd.Flags().GetString("crypted")
	if cfilename != "" {
		pc.CryptedFile = cfilename
	}
	// check for plaintext file option
	pfilename, _ := cmd.Flags().GetString("plaintext")
	if pfilename != "" {
		pc.PlainTextFile = pfilename
	}
	log.Debugf("decrypt file '%s' with method %s", pc.CryptedFile, pc.Method)

	log.Debugf("create plaintext file '%s'", pc.PlainTextFile)

	// check for keypass file option
	kp, _ := cmd.Flags().GetString("keypass")
	switch {
	case kp != "":
		pc.KeyPass = kp
		log.Debug("decrypt: keypass source: --keypass flag")
	case pc.KeyPass != "":
		log.Debug("decrypt: keypass source: config/env/default")
	default:
		log.Debug("decrypt: keypass source: none")
	}
	if err := checkKMSParams(); err != nil {
		return err
	}
	// do decrypt with default key
	lines, err := pc.DecryptFile()
	if err != nil && kp == "" && methodUsesKeypass(method) {
		if pw, _ := promptKeypass("Key passphrase"); pw != "" {
			pc.KeyPass = pw
			log.Debug("decrypt: keypass source: interactive prompt")
			lines, err = pc.DecryptFile()
		}
	}
	if err != nil {
		log.Errorf("decrypt failed: %s", err)
		return err
	}
	// write lines to file
	err = common.WriteStringToFile(pc.PlainTextFile, strings.Join(lines, "\n"))
	if err == nil {
		log.Infof("plaintext file '%s' successfully created", pc.PlainTextFile)
		fmt.Println("DONE")
	}
	return err
}
func init() {
	RootCmd.AddCommand(decryptCmd)
	// don't have variables populated here
	decryptCmd.PersistentFlags().StringP("plaintext", "t", "", "alternate plaintext file")
	decryptCmd.PersistentFlags().StringP("crypted", "c", "", "alternate crypted file")
	decryptCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
	decryptCmd.Flags().StringVar(&kmsKeyID, "kms_keyid", kmsKeyID, "KMS KeyID")
	decryptCmd.Flags().StringVar(&kmsEndpoint, "kms_endpoint", kmsEndpoint, "KMS Endpoint Url")
}
