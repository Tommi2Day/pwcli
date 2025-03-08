// Package cmd Commands
package cmd

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/common"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt plaintext file",
	Long: `Encrypt a plain file given in -p and saved as crypted file given by -c flag using given method.
default for plaintext File is <app>.plain and for crypted file is <app.pw>`,
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
	if method == typeKMS {
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
	}

	// make sure target directory exists
	dataDir := path.Dir(pc.CryptedFile)
	if dataDir != pc.DataDir {
		log.Infof("data directory %s differs from default %s", dataDir, pc.DataDir)
	}
	log.Debugf("data directory %s", dataDir)
	if !common.IsDir(dataDir) {
		log.Debugf("data directory %s doesnt exist", dataDir)
		err := os.MkdirAll(dataDir, 0700)
		if err != nil {
			log.Errorf("failed to create data directory %s: %s, choose anpther DataDir using -D", dataDir, err)
			return err
		}
		log.Infof("created data directory %s", dataDir)
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
	encryptCmd.Flags().StringVar(&kmsKeyID, "kms_keyid", kmsKeyID, "KMS KeyID")
	encryptCmd.Flags().StringVar(&kmsEndpoint, "kms_endpoint", kmsEndpoint, "KMS Endpoint Url")
}
