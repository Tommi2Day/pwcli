package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var awssmSecretID string
var awssmEndpoint = os.Getenv("SECRETSMANAGER_ENDPOINT")
var awssmKMSKeyID = os.Getenv("SECRETSMANAGER_KMS_KEY_ID")

var awssmCmd = &cobra.Command{
	Use:   typeAWSSM,
	Short: "handle AWS Secrets Manager functions",
	Long:  `Allows list, read and write AWS Secrets Manager secrets`,
}

var awssmReadCmd = &cobra.Command{
	Use:   cmdRead,
	Short: "read an AWS Secrets Manager secret",
	Long: `
read a secret from given secret ID/ARN, decoded as a JSON object
list all data below the secret in list_password syntax or give a key as extra arg to return only this value`,
	RunE:         awssmRead,
	SilenceUsage: true,
}
var awssmListCmd = &cobra.Command{
	Use:          "secrets",
	Aliases:      []string{cmdList, "ls"},
	Short:        "list secrets",
	Long:         `list the names of all secrets available in AWS Secrets Manager`,
	RunE:         awssmList,
	SilenceUsage: true,
}

var awssmWriteCmd = &cobra.Command{
	Use:          cmdWrite,
	Short:        "write json to an AWS Secrets Manager secret",
	Long:         `write a secret with json encoded data, creating it if it does not exist yet`,
	RunE:         awssmWrite,
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		f, e := cmd.Flags().GetString("data_file")
		if len(args) < 1 && (f == "" || e != nil) {
			return fmt.Errorf("requires data to write as second argument or 'data_file' set")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(awssmCmd)

	awssmCmd.PersistentFlags().StringVarP(&awssmSecretID, "path", "P", "", "AWS Secrets Manager secret ID/ARN to Read/Write")
	awssmCmd.PersistentFlags().StringVar(&awssmEndpoint, "awssm_endpoint", awssmEndpoint, "SECRETSMANAGER_ENDPOINT Url")

	hideGlobalFlags(awssmReadCmd)
	awssmReadCmd.Flags().BoolVarP(&jsonOut, "json", "J", false, "output as json")
	awssmReadCmd.Flags().BoolVarP(&exportOut, "export", "E", false, "output as bash export")
	awssmReadCmd.Flags().BoolVarP(&dotenvOut, "dotenv", "N", false,
		"output as KEY=value lines without 'export ' prefix, suitable for docker compose .env files")

	hideGlobalFlags(awssmWriteCmd)
	awssmWriteCmd.Flags().String("data_file", "", "Path to the json encoded file with the data to read from")
	awssmWriteCmd.Flags().StringVar(&awssmKMSKeyID, "awssm_kms_keyid", awssmKMSKeyID,
		"customer managed KMS key (ID, ARN or alias) to encrypt the secret with, default aws/secretsmanager (SECRETSMANAGER_KMS_KEY_ID)")

	hideGlobalFlags(awssmListCmd)
	awssmCmd.AddCommand(awssmReadCmd)
	awssmCmd.AddCommand(awssmWriteCmd)
	awssmCmd.AddCommand(awssmListCmd)
}

// setSecretsManagerEndpoint propagates the --awssm_endpoint flag / SECRETSMANAGER_ENDPOINT
// env var to the environment so pwlib.ConnectToSecretsManager picks it up.
func setSecretsManagerEndpoint() {
	if awssmEndpoint != "" {
		_ = os.Setenv("SECRETSMANAGER_ENDPOINT", awssmEndpoint)
	}
}

func awssmRead(_ *cobra.Command, args []string) error {
	log.Debugf("Secrets Manager Read entered for secret '%s'", awssmSecretID)

	if err := validateOutputFormats(); err != nil {
		return err
	}

	key := extractKey(args)
	setSecretsManagerEndpoint()
	svc, err := pwlib.ConnectToSecretsManager()
	if err != nil {
		return err
	}

	data, err := pwlib.SecretsManagerReadJSON(svc, awssmSecretID)
	if err != nil {
		return err
	}
	if data == nil {
		return fmt.Errorf("no entries returned")
	}
	log.Debug("Secrets Manager Read OK")
	if err = printSecretData(data, key, awssmSecretID); err != nil {
		return err
	}
	log.Info("Secrets Manager Data successfully processed")
	return nil
}

func awssmWrite(cmd *cobra.Command, args []string) error {
	var (
		err      error
		datafile string
		content  string
		data     map[string]interface{}
	)

	log.Debug("Secrets Manager Write entered")
	if len(args) > 0 {
		content = args[0]
	} else {
		datafile, err = cmd.Flags().GetString("data_file")
		if err == nil {
			content, err = common.ReadFileToString(datafile)
			if err != nil {
				err = fmt.Errorf("could not read data file '%s': %s", datafile, err)
				return err
			}
		} else {
			return err
		}
	}
	if len(content) < 3 {
		err = fmt.Errorf("no input to write, use 'data_file' as file or Arg0 as string with json data")
		return err
	}
	err = json.Unmarshal([]byte(content), &data)
	if err != nil {
		err = fmt.Errorf("could not unmarshal json data: %s", err)
		return err
	}

	setSecretsManagerEndpoint()
	if awssmKMSKeyID != "" {
		log.Debugf("use KMS key %s for Secrets Manager write", awssmKMSKeyID)
		_ = os.Setenv("SECRETSMANAGER_KMS_KEY_ID", awssmKMSKeyID)
	}
	svc, err := pwlib.ConnectToSecretsManager()
	if err != nil {
		return err
	}
	err = pwlib.SecretsManagerWrite(svc, awssmSecretID, data)
	if err == nil {
		log.Info("Secrets Manager Write OK")
		fmt.Println("OK")
	}
	return err
}

func awssmList(_ *cobra.Command, _ []string) error {
	log.Debug("Secrets Manager list entered")
	setSecretsManagerEndpoint()
	svc, err := pwlib.ConnectToSecretsManager()
	if err != nil {
		return err
	}

	secrets, err := pwlib.SecretsManagerList(svc)
	if err != nil {
		return fmt.Errorf("list command failed:%s", err)
	}
	log.Infof("Secrets Manager List returned %d entries", len(secrets))
	for _, s := range secrets {
		log.Debugf("Secrets Manager List entry: %s", s)
		fmt.Println(s)
	}
	return nil
}
