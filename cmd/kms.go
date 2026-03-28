package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
)

const aliasPrefix = "alias/"

var kmsKeyID = common.GetStringEnv("KMS_KEYID", "")
var kmsEndpoint = common.GetStringEnv("KMS_ENDPOINT", "")

var kmsCmd = &cobra.Command{
	Use:   "kms",
	Short: "KMS key management",
	Long:  `Manage KMS keys: generate, describe, export, delete and handle policies.`,
}

var kmsGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new KMS key",
	RunE:  kmsGenerate,
}

var kmsDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a KMS key",
	RunE:  kmsDescribe,
}

var kmsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a KMS public key",
	RunE:  kmsExport,
}

var kmsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Schedule deletion of a KMS key",
	RunE:  kmsDelete,
}

var kmsGetPolicyCmd = &cobra.Command{
	Use:   "get-policy",
	Short: "Get KMS key policy",
	RunE:  kmsGetPolicy,
}

var kmsPutPolicyCmd = &cobra.Command{
	Use:   "put-policy",
	Short: "Put KMS key policy",
	RunE:  kmsPutPolicy,
}

var kmsCreateAliasCmd = &cobra.Command{
	Use:   "create-alias",
	Short: "Create an alias for a KMS key",
	RunE:  kmsCreateAlias,
}

var kmsUpdateAliasCmd = &cobra.Command{
	Use:   "update-alias",
	Short: "Update an existing KMS key alias to point to a different key",
	RunE:  kmsUpdateAlias,
}

var kmsDeleteAliasCmd = &cobra.Command{
	Use:   "delete-alias",
	Short: "Delete a KMS key alias",
	RunE:  kmsDeleteAlias,
}

var kmsListAliasesCmd = &cobra.Command{
	Use:   "list-aliases",
	Short: "List all KMS key aliases",
	RunE:  kmsListAliases,
}

func init() {
	RootCmd.AddCommand(kmsCmd)
	for _, sub := range []*cobra.Command{
		kmsGenerateCmd, kmsDescribeCmd, kmsExportCmd, kmsDeleteCmd,
		kmsGetPolicyCmd, kmsPutPolicyCmd,
		kmsCreateAliasCmd, kmsUpdateAliasCmd, kmsDeleteAliasCmd, kmsListAliasesCmd,
	} {
		hideGlobalFlags(sub, "no-prompt")
		kmsCmd.AddCommand(sub)
	}

	kmsCmd.PersistentFlags().StringVar(&kmsKeyID, "kms_keyid", kmsKeyID, "KMS KeyID")
	kmsCmd.PersistentFlags().StringVar(&kmsEndpoint, "kms_endpoint", kmsEndpoint, "KMS Endpoint Url")

	kmsGenerateCmd.Flags().StringP("description", "d", "", "Key description")
	kmsGenerateCmd.Flags().StringP("type", "t", "symmetric", "Key type (symmetric|rsa)")
	kmsDeleteCmd.Flags().Int32P("days", "w", 7, "Waiting days before deletion (7-30)")
	kmsPutPolicyCmd.Flags().StringP("file", "f", "", "Policy file (JSON)")
	kmsCreateAliasCmd.Flags().StringP("alias", "n", "", "Alias name (must start with alias/)")
	kmsUpdateAliasCmd.Flags().StringP("alias", "n", "", "Alias name (must start with alias/)")
	kmsDeleteAliasCmd.Flags().StringP("alias", "n", "", "Alias name (must start with alias/)")
	kmsListAliasesCmd.Flags().StringP("key-id", "k", "", "Filter aliases by key ID (optional)")
}

func getKMSClient() (*kms.Client, error) {
	if kmsEndpoint != "" {
		_ = os.Setenv("KMS_ENDPOINT", kmsEndpoint)
	}
	svc := pwlib.ConnectToKMS()
	if svc == nil {
		return nil, fmt.Errorf("failed to connect to KMS")
	}
	return svc, nil
}

func checkKMSKeyID(cmd *cobra.Command) error {
	var keyIDToUse string
	if cmd != nil {
		if val, err := cmd.Flags().GetString("kms_keyid"); err == nil && val != "" {
			keyIDToUse = val
		}
	}

	if keyIDToUse == "" {
		if kmsKeyID != "" {
			keyIDToUse = kmsKeyID
		} else {
			keyIDToUse = common.GetStringEnv("KMS_KEYID", "")
		}
	}

	if keyIDToUse == "" {
		return fmt.Errorf("need parameter kms_keyid to proceed")
	}

	kmsKeyID = keyIDToUse
	return nil
}

func kmsGenerate(cmd *cobra.Command, _ []string) error {
	svc, err := getKMSClient()
	if err != nil {
		return err
	}
	desc, _ := cmd.Flags().GetString("description")
	keytype, _ := cmd.Flags().GetString("type")
	spec := types.CustomerMasterKeySpecSymmetricDefault
	if keytype == pwlib.KeyTypeRSA {
		spec = types.CustomerMasterKeySpecRsa2048
	}
	output, err := pwlib.GenKMSKey(svc, string(spec), desc, nil)
	if err != nil {
		return err
	}
	keyID, _ := pwlib.GetKMSKeyIDs(output.KeyMetadata)
	log.Infof("KMS key generated: %s", keyID)
	cmd.Printf("KeyID: %s\n", keyID)
	return nil
}

func kmsDescribe(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}
	output, err := pwlib.DescribeKMSKey(svc, kmsKeyID)
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(output.KeyMetadata, "", "  ")
	cmd.Println(string(data))
	return nil
}

func kmsExport(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}

	// Get key metadata to determine key type
	keyMetadata, err := svc.DescribeKey(context.TODO(), &kms.DescribeKeyInput{
		KeyId: aws.String(kmsKeyID),
	})
	if err != nil {
		return fmt.Errorf("failed to describe key: %w", err)
	}

	keySpec := keyMetadata.KeyMetadata.KeySpec
	keyUsage := keyMetadata.KeyMetadata.KeyUsage

	// Handle RSA keys - export public key
	if keySpec == types.KeySpecRsa2048 || keySpec == types.KeySpecRsa3072 || keySpec == types.KeySpecRsa4096 {
		return kmsExportRSAKey(cmd, svc, keyMetadata)
	}

	// Handle Symmetric keys - generate data key
	if keySpec == types.KeySpecSymmetricDefault {
		return kmsExportSymmetricKey(cmd, svc, keyUsage)
	}

	return fmt.Errorf("unsupported key type: %s", keySpec)
}

func kmsExportRSAKey(cmd *cobra.Command, svc *kms.Client, keyMetadata *kms.DescribeKeyOutput) error {
	// Get public key
	publicKeyOutput, err := svc.GetPublicKey(context.TODO(), &kms.GetPublicKeyInput{
		KeyId: aws.String(kmsKeyID),
	})
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Export public key to file
	publicKeyFile := fmt.Sprintf("%s.pub", kmsKeyID)
	err = os.WriteFile(publicKeyFile, publicKeyOutput.PublicKey, 0600)
	if err != nil {
		return fmt.Errorf("failed to write public key file: %w", err)
	}

	keyUsage := keyMetadata.KeyMetadata.KeyUsage

	// Only generate data key for ENCRYPT_DECRYPT usage keys
	var encryptedKeyFile, plaintextKeyFile string
	if keyUsage == types.KeyUsageTypeEncryptDecrypt {
		// Generate a data key to demonstrate key wrapping
		dataKeyOutput, err := svc.GenerateDataKey(context.TODO(), &kms.GenerateDataKeyInput{
			KeyId:   aws.String(kmsKeyID),
			KeySpec: types.DataKeySpecAes256,
		})
		if err != nil {
			return fmt.Errorf("failed to generate data key: %w", err)
		}

		// Export encrypted data key to file (simulates private key wrapping)
		encryptedKeyFile = fmt.Sprintf("%s.key.encrypted", kmsKeyID)
		err = os.WriteFile(encryptedKeyFile, dataKeyOutput.CiphertextBlob, 0600)
		if err != nil {
			return fmt.Errorf("failed to write encrypted key file: %w", err)
		}

		// Export plaintext data key to file (for demonstration purposes)
		plaintextKeyFile = fmt.Sprintf("%s.key.plaintext", kmsKeyID)
		err = os.WriteFile(plaintextKeyFile, dataKeyOutput.Plaintext, 0600)
		if err != nil {
			return fmt.Errorf("failed to write plaintext key file: %w", err)
		}

		log.Infof("RSA key exported: public key to %s, encrypted key to %s, plaintext key to %s",
			publicKeyFile, encryptedKeyFile, plaintextKeyFile)
		cmd.Printf("✓ RSA Key Export Complete\n")
		cmd.Printf("  Public Key:      %s\n", publicKeyFile)
		cmd.Printf("  Encrypted Key:   %s\n", encryptedKeyFile)
		cmd.Printf("  Plaintext Key:   %s\n", plaintextKeyFile)
	} else {
		// For SIGN_VERIFY keys, only export public key
		log.Infof("RSA signing key exported: public key to %s", publicKeyFile)
		cmd.Printf("✓ RSA Key Export Complete\n")
		cmd.Printf("  Public Key:      %s\n", publicKeyFile)
		cmd.Printf("  Note: Signing keys (SIGN_VERIFY) do not support data key generation\n")
	}

	cmd.Printf("  Key ID:          %s\n", kmsKeyID)
	cmd.Printf("  Key Spec:        %s\n", keyMetadata.KeyMetadata.KeySpec)
	cmd.Printf("  Key Usage:       %s\n", keyUsage)

	return nil
}

func kmsExportSymmetricKey(cmd *cobra.Command, svc *kms.Client, keyUsage types.KeyUsageType) error {
	// Generate data key for symmetric key export
	dataKeyOutput, err := svc.GenerateDataKey(context.TODO(), &kms.GenerateDataKeyInput{
		KeyId:   aws.String(kmsKeyID),
		KeySpec: types.DataKeySpecAes256,
	})
	if err != nil {
		return fmt.Errorf("failed to generate data key: %w", err)
	}

	// Export plaintext data key to file
	plaintextKeyFile := fmt.Sprintf("%s.key.plaintext", kmsKeyID)
	err = os.WriteFile(plaintextKeyFile, dataKeyOutput.Plaintext, 0600)
	if err != nil {
		return fmt.Errorf("failed to write plaintext key file: %w", err)
	}

	// Export encrypted data key to file
	encryptedKeyFile := fmt.Sprintf("%s.key.encrypted", kmsKeyID)
	err = os.WriteFile(encryptedKeyFile, dataKeyOutput.CiphertextBlob, 0600)
	if err != nil {
		return fmt.Errorf("failed to write encrypted key file: %w", err)
	}

	// Export key metadata to JSON file
	metadataFile := fmt.Sprintf("%s.metadata.json", kmsKeyID)
	metadata := map[string]interface{}{
		"KeyId":        kmsKeyID,
		"KeyUsage":     keyUsage,
		"DataKeySize":  len(dataKeyOutput.Plaintext),
		"EncryptedKey": len(dataKeyOutput.CiphertextBlob),
	}
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	err = os.WriteFile(metadataFile, metadataJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	log.Infof("Symmetric key exported: plaintext to %s, encrypted to %s, metadata to %s",
		plaintextKeyFile, encryptedKeyFile, metadataFile)
	cmd.Printf("✓ Symmetric Key Export Complete\n")
	cmd.Printf("  Plaintext Key:   %s\n", plaintextKeyFile)
	cmd.Printf("  Encrypted Key:   %s\n", encryptedKeyFile)
	cmd.Printf("  Metadata:        %s\n", metadataFile)
	cmd.Printf("  Key ID:          %s\n", kmsKeyID)
	cmd.Printf("  Key Usage:       %s\n", keyUsage)

	return nil
}

func kmsDelete(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}
	days, _ := cmd.Flags().GetInt32("days")
	output, err := svc.ScheduleKeyDeletion(context.TODO(), &kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(kmsKeyID),
		PendingWindowInDays: aws.Int32(days),
	})
	if err != nil {
		return err
	}
	log.Infof("KMS key %s scheduled for deletion at %v", kmsKeyID, output.DeletionDate)
	cmd.Printf("DeletionDate: %v\n", output.DeletionDate)
	return nil
}

func kmsGetPolicy(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}
	output, err := svc.GetKeyPolicy(context.TODO(), &kms.GetKeyPolicyInput{
		KeyId:      aws.String(kmsKeyID),
		PolicyName: aws.String("default"),
	})
	if err != nil {
		return err
	}
	cmd.Println(*output.Policy)
	return nil
}

func kmsPutPolicy(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}
	file, _ := cmd.Flags().GetString("file")
	if file == "" {
		return fmt.Errorf("policy file is required")
	}
	policy, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = svc.PutKeyPolicy(context.TODO(), &kms.PutKeyPolicyInput{
		KeyId:      aws.String(kmsKeyID),
		PolicyName: aws.String("default"),
		Policy:     aws.String(string(policy)),
	})
	if err != nil {
		return err
	}
	log.Infof("KMS key policy updated for %s", kmsKeyID)
	cmd.Println("DONE")
	return nil
}

func kmsCreateAlias(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}

	aliasName, _ := cmd.Flags().GetString("alias")
	if aliasName == "" {
		return fmt.Errorf("alias name is required")
	}

	// Ensure alias starts with "alias/"
	if len(aliasName) < len(aliasPrefix) || aliasName[:len(aliasPrefix)] != aliasPrefix {
		return fmt.Errorf("alias name must start with 'alias/'")
	}

	_, err = svc.CreateAlias(context.TODO(), &kms.CreateAliasInput{
		AliasName:   aws.String(aliasName),
		TargetKeyId: aws.String(kmsKeyID),
	})
	if err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	log.Infof("KMS alias '%s' created for key %s", aliasName, kmsKeyID)
	cmd.Printf("Alias '%s' created successfully for key %s\n", aliasName, kmsKeyID)
	return nil
}

func kmsUpdateAlias(cmd *cobra.Command, _ []string) error {
	if err := checkKMSKeyID(cmd); err != nil {
		return err
	}
	svc, err := getKMSClient()
	if err != nil {
		return err
	}

	aliasName, _ := cmd.Flags().GetString("alias")
	if aliasName == "" {
		return fmt.Errorf("alias name is required")
	}

	// Ensure alias starts with "alias/"
	if len(aliasName) < len(aliasPrefix) || aliasName[:len(aliasPrefix)] != aliasPrefix {
		return fmt.Errorf("alias name must start with 'alias/'")
	}

	_, err = svc.UpdateAlias(context.TODO(), &kms.UpdateAliasInput{
		AliasName:   aws.String(aliasName),
		TargetKeyId: aws.String(kmsKeyID),
	})
	if err != nil {
		return fmt.Errorf("failed to update alias: %w", err)
	}

	log.Infof("KMS alias '%s' updated to point to key %s", aliasName, kmsKeyID)
	cmd.Printf("Alias '%s' updated successfully to point to key %s\n", aliasName, kmsKeyID)
	return nil
}

func kmsDeleteAlias(cmd *cobra.Command, _ []string) error {
	svc, err := getKMSClient()
	if err != nil {
		return err
	}

	aliasName, _ := cmd.Flags().GetString("alias")
	if aliasName == "" {
		return fmt.Errorf("alias name is required")
	}

	// Ensure alias starts with "alias/"
	if len(aliasName) < len(aliasPrefix) || aliasName[:len(aliasPrefix)] != aliasPrefix {
		return fmt.Errorf("alias name must start with 'alias/'")
	}

	_, err = svc.DeleteAlias(context.TODO(), &kms.DeleteAliasInput{
		AliasName: aws.String(aliasName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete alias: %w", err)
	}

	log.Infof("KMS alias '%s' deleted successfully", aliasName)
	cmd.Printf("Alias '%s' deleted successfully\n", aliasName)
	return nil
}

func kmsListAliases(cmd *cobra.Command, _ []string) error {
	svc, err := getKMSClient()
	if err != nil {
		return err
	}

	filterKeyID, _ := cmd.Flags().GetString("key-id")

	input := &kms.ListAliasesInput{}
	if filterKeyID != "" {
		input.KeyId = aws.String(filterKeyID)
	}

	paginator := kms.NewListAliasesPaginator(svc, input)

	aliasCount := 0
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to list aliases: %w", err)
		}

		for _, alias := range output.Aliases {
			if alias.AliasName != nil {
				targetKey := "N/A"
				if alias.TargetKeyId != nil {
					targetKey = *alias.TargetKeyId
				}
				cmd.Printf("%-40s -> %s\n", *alias.AliasName, targetKey)
				aliasCount++
			}
		}
	}

	log.Infof("Listed %d KMS alias(es)", aliasCount)
	if aliasCount == 0 {
		cmd.Println("No aliases found")
	}
	return nil
}
