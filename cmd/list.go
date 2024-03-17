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
	listCmd.Flags().StringVar(&kmsKeyID, "kms_keyid", kmsKeyID, "KMS KeyID")
	listCmd.Flags().StringVar(&kmsEndpoint, "kms_endpoint", kmsEndpoint, "KMS Endpoint Url")
}
