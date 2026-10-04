package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tommi2day/gomodules/common"

	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/pwlib"
)

var vaultToken = os.Getenv("VAULT_TOKEN")
var vaultAddr = os.Getenv("VAULT_ADDR")
var vaultPath string
var logical bool
var kvMount = "secret/"
var jsonOut = false
var exportOut = false
var dotenvOut = false
var vaultCmd = &cobra.Command{
	Use:   typeVault,
	Short: "handle vault functions",
	Long:  `Allows list, read and write vault secrets`,
}

var vaultReadCmd = &cobra.Command{
	Use:   cmdRead,
	Short: "read a vault secret",
	Long: `
read a secret from given path in KV2 or Logical mode
list all data below path in list_password syntax or give a key as extra arg to return only this value`,
	RunE:         vaultRead,
	SilenceUsage: true,
}
var vaultListCmd = &cobra.Command{
	Use:          "secrets",
	Aliases:      []string{cmdList, "ls"},
	Short:        "list secrets",
	Long:         `list secrets recursive below given path (without content)`,
	RunE:         vaultList,
	SilenceUsage: true,
}

var vaultWriteCmd = &cobra.Command{
	Use:          cmdWrite,
	Short:        "write json to vault path",
	Long:         `write a secret to given path in KV2 or Logical mode with json encoded data`,
	RunE:         vaultWrite,
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
	RootCmd.AddCommand(vaultCmd)

	vaultCmd.PersistentFlags().StringVar(&vaultAddr, "vault_addr", vaultAddr, "VAULT_ADDR Url")
	vaultCmd.PersistentFlags().StringVar(&vaultToken, "vault_token", vaultToken, "VAULT_TOKEN")
	vaultCmd.PersistentFlags().StringVarP(&vaultPath, "path", "P", "", "Vault secret Path to Read/Write/List")
	vaultCmd.PersistentFlags().BoolVarP(&logical, "logical", "L", false, "Use Logical Api, default is KV2")
	vaultCmd.PersistentFlags().StringVarP(&kvMount, "mount", "M", kvMount, "Mount Path of the Secret engine")

	hideGlobalFlags(vaultReadCmd, "no-prompt")
	vaultReadCmd.Flags().BoolVarP(&jsonOut, "json", "J", false, "output as json")
	vaultReadCmd.Flags().BoolVarP(&exportOut, "export", "E", false, "output as bash export")
	vaultReadCmd.Flags().BoolVarP(&dotenvOut, "dotenv", "N", false,
		"output as KEY=value lines without 'export ' prefix, suitable for docker compose .env files")

	hideGlobalFlags(vaultWriteCmd, "no-prompt")
	vaultWriteCmd.Flags().String("data_file", "", "Path to the json encoded file with the data to read from")

	hideGlobalFlags(vaultListCmd, "no-prompt")
	vaultCmd.AddCommand(vaultReadCmd)
	vaultCmd.AddCommand(vaultWriteCmd)
	vaultCmd.AddCommand(vaultListCmd)
}

func vaultRead(_ *cobra.Command, args []string) error {
	log.Debugf("Vault Read entered for path '%s'", vaultPath)

	if err := validateOutputFormats(); err != nil {
		return err
	}

	key := extractKey(args)
	vc, _ := pwlib.VaultConfig(vaultAddr, vaultToken)

	err := readVaultData(vc, key)
	if err == nil {
		log.Info("Vault Data successfully processed")
	}
	return err
}

func validateOutputFormats() error {
	selected := 0
	for _, f := range []bool{jsonOut, exportOut, dotenvOut} {
		if f {
			selected++
		}
	}
	if selected > 1 {
		return fmt.Errorf("cannot use more than one of 'json', 'export' and 'dotenv' output format at the same time")
	}
	return nil
}

func extractKey(args []string) string {
	if len(args) > 0 {
		key := args[0]
		log.Debugf("Selected Key '%s'", key)
		return key
	}
	return ""
}

func readVaultData(vc *vault.Client, key string) error {
	if logical {
		return readVaultDataLogical(vc, key)
	}
	return readVaultDataKV(vc, key)
}

func readVaultDataLogical(vc *vault.Client, key string) error {
	vs, err := pwlib.VaultRead(vc, vaultPath)
	if err != nil {
		return err
	}

	if vs == nil {
		return fmt.Errorf("no entries returned")
	}

	log.Debug("Vault Read OK")
	return printSecretData(vs.Data, key, vaultPath)
}

func readVaultDataKV(vc *vault.Client, key string) error {
	kvs, err := pwlib.VaultKVRead(vc, kvMount, vaultPath)
	if err != nil {
		return err
	}

	if kvs == nil {
		return fmt.Errorf("no entries returned")
	}

	log.Debug("Vault KVRead OK")
	return printSecretData(kvs.Data, key, vaultPath)
}

func vaultWrite(cmd *cobra.Command, args []string) error {
	var (
		err       error
		datafile  string
		content   string
		vaultData map[string]interface{}
		vc        *vault.Client
	)

	log.Debug("Vault Write entered")
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
	err = json.Unmarshal([]byte(content), &vaultData)
	if err != nil {
		err = fmt.Errorf("could not unmarshal json data: %s", err)
		return err
	}
	vc, _ = pwlib.VaultConfig(vaultAddr, vaultToken)
	if logical {
		log.Debug("Write data with logical api")
		err = pwlib.VaultWrite(vc, vaultPath, vaultData)
	} else {
		log.Debug("Write data with KV api")
		err = pwlib.VaultKVWrite(vc, kvMount, vaultPath, vaultData)
	}
	if err == nil {
		log.Info("Vault Write OK")
		fmt.Println("OK")
	}
	return err
}

func vaultList(_ *cobra.Command, _ []string) error {
	var (
		err     error
		vc      *vault.Client
		secrets []string
	)

	log.Debugf("Vault list entered for path '%s'", vaultPath)
	vc, _ = pwlib.VaultConfig(vaultAddr, vaultToken)
	vp := vaultPath
	mount := kvMount
	secrets, err = pwlib.VaultList(vc, mount, vp)
	l := len(secrets)
	if err == nil {
		log.Infof("Vault List returned %d entries", l)
		for _, k := range secrets {
			k = strings.TrimPrefix(k, mount)
			k = strings.TrimPrefix(k, "/metadata/")
			log.Debugf("Vault List entry: %s", k)
			fmt.Println(k)
		}
	} else {
		err = fmt.Errorf("list command failed:%s", err)
	}
	return err
}

// printSecretData prints the given secret data map, either a single key's value or the
// full data set. path identifies the secret's origin (vault path / AWS secret ID) and is
// only used to build the "system" column of the default output format.
func printSecretData(secretData map[string]interface{}, key string, path string) (err error) {
	if key != "" {
		return printSingleKey(secretData, key)
	}

	if len(secretData) == 0 {
		return fmt.Errorf("no data found")
	}

	return printAllData(secretData, path)
}

func printSingleKey(vaultData map[string]interface{}, key string) error {
	value, ok := vaultData[key].(string)
	if !ok {
		return fmt.Errorf("key '%s' not found", key)
	}
	log.Debugf("READ:%s", value)
	fmt.Printf("%s", value)
	return nil
}

func printAllData(vaultData map[string]interface{}, path string) error {
	switch {
	case jsonOut:
		return printJSONOutput(vaultData)
	case exportOut:
		printExportOutput(vaultData)
		return nil
	case dotenvOut:
		printDotenvOutput(vaultData)
		return nil
	default:
		printDefaultOutput(vaultData, path)
		return nil
	}
}

func printJSONOutput(vaultData map[string]interface{}) error {
	jsonData, err := json.Marshal(vaultData)
	if err != nil {
		return fmt.Errorf("cannot generate json output:%s", err)
	}
	log.Debugf("JSON:\n%s", jsonData)
	fmt.Printf("%s\n", jsonData)
	return nil
}

func shellEscapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

// sanitizeExportKey turns a vault key into a safe shell variable name so that
// injected characters (spaces, quotes, ';', '`', '$', newlines, ...) cannot
// break out of the "export NAME=..." statement when the output is eval'd.
func sanitizeExportKey(k string) string {
	upper := strings.ToUpper(k)
	sanitized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			return r
		default:
			return '_'
		}
	}, upper)
	if sanitized == "" || (sanitized[0] >= '0' && sanitized[0] <= '9') {
		sanitized = "_" + sanitized
	}
	return sanitized
}

func printExportOutput(vaultData map[string]interface{}) {
	printKeyEqualsValueOutput(vaultData, "export ")
}

// printDotenvOutput writes KEY='value' lines without the 'export ' prefix, so the
// output can be used directly as a docker compose .env file.
func printDotenvOutput(vaultData map[string]interface{}) {
	printKeyEqualsValueOutput(vaultData, "")
}

func printKeyEqualsValueOutput(vaultData map[string]interface{}, prefix string) {
	for k, v := range vaultData {
		escaped := shellEscapeSingleQuote(fmt.Sprintf("%v", v))
		o := fmt.Sprintf("%s%s='%s'\n", prefix, sanitizeExportKey(k), escaped)
		log.Debugf("OUTPUT:\n%s", o)
		fmt.Printf("%s", o)
	}
}

func printDefaultOutput(vaultData map[string]interface{}, path string) {
	sysKey := strings.ReplaceAll(path, ":", "_")
	for k, v := range vaultData {
		o := fmt.Sprintf("%s:%s:%v\n", sysKey, k, v)
		log.Debugf("READ:%s", o)
		fmt.Printf("%s", o)
	}
}
