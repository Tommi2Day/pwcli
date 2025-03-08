// Package cmd Commands
package cmd

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
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
	var err error
	if method == "" {
		method = typeOpenSSL
	}
	log.Debugf("genkey for method %s", method)
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate key password '%s'", kp)
	}
	switch method {
	case typeGO, typeOpenSSL:
		// make sure target directory exists
		keyDir := path.Dir(pc.PrivateKeyFile)
		if keyDir != pc.DataDir {
			log.Infof("taget key directory %s differs from default %s", keyDir, pc.KeyDir)
		}
		log.Debugf("key directory %s", keyDir)
		if !common.IsDir(keyDir) {
			log.Debugf("key directory %s doesnt exist", keyDir)
			err = os.MkdirAll(keyDir, 0700)
			if err != nil {
				log.Errorf("failed to create key directory %s: %s, choose another one using -K", keyDir, err)
				return err
			}
			log.Infof("created key directory %s", keyDir)
		}
		// genkey
		_, _, err = pwlib.GenRsaKey(pc.PubKeyFile, pc.PrivateKeyFile, pc.KeyPass)
		if err == nil {
			log.Infof("New key pair generated as %s and %s", pc.PubKeyFile, pc.PrivateKeyFile)
			fmt.Println("DONE")
		}
		return err
	}
	return fmt.Errorf("method %s doesnt support keygeneration", method)
}

func init() {
	RootCmd.AddCommand(generateCmd)
	// don't have variables populated here
	generateCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
}
