// Package cmd Commands
package cmd

import (
	"fmt"
	"os"
	"path"

	"filippo.io/age"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

const defaultKeyType = "rsa"

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate a new Keypair",
	Long: `Generates a new pair of keys (ecdsa, rsa, age, gpg)
optionally you may assign an individual key password using -p flag
`,
	RunE:         genkey,
	SilenceUsage: true,
}

func genkey(cmd *cobra.Command, _ []string) error {
	var err error
	kp, _ := cmd.Flags().GetString("keypass")
	keytype, _ := cmd.Flags().GetString("type")
	if kp != "" {
		pc.KeyPass = kp
		log.Debugf("use alternate key password '%s'", kp)
	}
	if keytype == "" {
		keytype = pc.KeyType
	} else {
		switch keytype {
		case pwlib.KeyTypeAGE:
			pc.Method = typeAGE
		case pwlib.KeyTypeGPG:
			pc.Method = typeGPG
		}
	}
	pc = pwlib.NewConfig(pc.AppName, pc.DataDir, pc.DataDir, pc.KeyPass, pc.Method)
	// make sure target directory exists
	keyDir := path.Dir(pc.PrivateKeyFile)
	if keyDir != pc.DataDir {
		log.Infof("target key directory %s differs from default %s", keyDir, pc.KeyDir)
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

	switch keytype {
	case pwlib.KeyTypeRSA:
		_, _, err = pwlib.GenRsaKey(pc.PubKeyFile, pc.PrivateKeyFile, pc.KeyPass)
	case pwlib.KeyTypeECDSA:
		_, _, err = pwlib.GenEcdsaKey(pc.PubKeyFile, pc.PrivateKeyFile, pc.KeyPass)
	case pwlib.KeyTypeAGE:
		var identity any
		identity, _, err = pwlib.CreateAgeIdentity()
		if err == nil {
			err = pwlib.ExportAgeKeyPair(identity.(*age.X25519Identity), pc.PubKeyFile, pc.PrivateKeyFile)
		}
	case pwlib.KeyTypeGPG:
		var entity any
		entity, _, err = pwlib.CreateGPGEntity(pc.AppName, "key for "+pc.AppName, pc.AppName+"@local", pc.KeyPass)
		if err == nil {
			err = pwlib.ExportGPGKeyPair(entity.(*openpgp.Entity), pc.PubKeyFile, pc.PrivateKeyFile)
		}
	default:
		return fmt.Errorf("key type %s not supported", pc.KeyType)
	}

	if err == nil {
		log.Infof("New key pair generated as %s and %s", pc.PubKeyFile, pc.PrivateKeyFile)
		cmd.Println("DONE")
	}
	return err
}

func init() {
	RootCmd.AddCommand(generateCmd)
	// don't have variables populated here
	generateCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
	generateCmd.Flags().StringP("type", "t", defaultKeyType, "key type: ecdsa|rsa|age|gpg")
}
