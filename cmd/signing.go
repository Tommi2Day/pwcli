package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/common"
)

var signKmsKeyID string
var signKmsEndpoint string

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a file",
	Long:  `Sign a file given in -t and saved as signature file given by -s flag using given method.`,
	RunE:  sign,
}

var verifyCmd = &cobra.Command{
	Use:     "verify",
	Aliases: []string{"vs"},
	Short:   "Verify a file signature",
	Long:    `Verify a file given in -t against a signature file given by -s flag using given method.`,
	RunE:    verify,
}

func checkKMSSignParams() error {
	if method == typeKMS {
		if signKmsKeyID == "" {
			signKmsKeyID = common.GetStringEnv("KMS_KEYID", "")
			log.Debugf("KMS KeyID from environment: '%s'", signKmsKeyID)
		}
		if signKmsKeyID == "" {
			return fmt.Errorf("need parameter kms_keyid to proceed")
		}
		if signKmsEndpoint != "" {
			log.Debugf("use KMS endpoint %s", signKmsEndpoint)
			_ = os.Setenv("KMS_ENDPOINT", signKmsEndpoint)
		}
		log.Debugf("use KMS method with keyid %s", signKmsKeyID)
		pc.KMSKeyID = signKmsKeyID
	}
	return nil
}

func sign(cmd *cobra.Command, _ []string) error {
	log.Debug("sign called")
	sfilename, _ := cmd.Flags().GetString("signature")
	if sfilename != "" {
		pc.SignatureFile = sfilename
	}
	pfilename, _ := cmd.Flags().GetString("plaintext")
	if pfilename != "" {
		pc.PlainTextFile = pfilename
	}
	kp, _ := cmd.Flags().GetString("keypass")
	if kp != "" {
		pc.KeyPass = kp
	}

	if err := checkKMSSignParams(); err != nil {
		return err
	}

	err := pc.SignFile()
	if err != nil {
		log.Errorf("sign failed: %s", err)
		return err
	}
	log.Infof("signature file '%s' successfully created", pc.SignatureFile)
	cmd.Println("DONE")
	return nil
}

func verify(cmd *cobra.Command, _ []string) error {
	log.Debug("verify called")
	sfilename, _ := cmd.Flags().GetString("signature")
	if sfilename != "" {
		pc.SignatureFile = sfilename
	}
	pfilename, _ := cmd.Flags().GetString("plaintext")
	if pfilename != "" {
		pc.PlainTextFile = pfilename
	}

	if err := checkKMSSignParams(); err != nil {
		return err
	}

	valid, err := pc.VerifyFile()
	if err != nil {
		log.Errorf("verify failed: %s", err)
		return err
	}
	if valid {
		log.Infof("signature for file '%s' is valid", pc.PlainTextFile)
		cmd.Println("VALID")
	} else {
		log.Errorf("signature for file '%s' is INVALID", pc.PlainTextFile)
		cmd.Println("INVALID")
	}
	return nil
}

func init() {
	RootCmd.AddCommand(signCmd)
	RootCmd.AddCommand(verifyCmd)

	signCmd.Flags().StringP("plaintext", "t", "", "alternate plaintext file")
	signCmd.Flags().StringP("signature", "s", "", "alternate signature file")
	signCmd.Flags().StringP("keypass", "p", "", "dedicated password for the private key")
	signCmd.Flags().StringVar(&signKmsKeyID, "kms_keyid", "", "KMS KeyID")
	signCmd.Flags().StringVar(&signKmsEndpoint, "kms_endpoint", "", "KMS Endpoint Url")

	verifyCmd.Flags().StringP("plaintext", "t", "", "alternate plaintext file")
	verifyCmd.Flags().StringP("signature", "s", "", "alternate signature file")
	verifyCmd.Flags().StringVar(&signKmsKeyID, "kms_keyid", "", "KMS KeyID")
	verifyCmd.Flags().StringVar(&signKmsEndpoint, "kms_endpoint", "", "KMS Endpoint Url")
}
