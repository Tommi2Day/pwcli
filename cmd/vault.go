package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

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
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "handle vault functions",
	Long:  `Allows list, read and write vault secrets`,
}

var vaultReadCmd = &cobra.Command{
	Use:   "read",
	Short: "read a vault secret",
	Long: `
read a secret from given path in KV2 or Logical mode
list all data below path in list_password syntax or give a key as extra arg to return only this value`,
	RunE:         vaultRead,
	SilenceUsage: true,
}
var vaultListCmd = &cobra.Command{
	Use:          "list",
	Short:        "list secrets",
	Long:         `list secrets one step below given path (without content)`,
	RunE:         vaultList,
	SilenceUsage: true,
}

var vaultWriteCmd = &cobra.Command{
	Use:          "write",
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
	vaultCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		// Hide flag for this command
		_ = command.Flags().MarkHidden("app")
		_ = command.Flags().MarkHidden("keydir")
		_ = command.Flags().MarkHidden("datadir")
		_ = command.Flags().MarkHidden("config")
		_ = command.Flags().MarkHidden("method")
		// Call parent help func
		command.Parent().HelpFunc()(command, strings)
	})
	RootCmd.AddCommand(vaultCmd)

	vaultCmd.PersistentFlags().StringVarP(&vaultAddr, "vault_addr", "A", vaultAddr, "VAULT_ADDR Url")
	vaultCmd.PersistentFlags().StringVarP(&vaultToken, "vault_token", "T", vaultToken, "VAULT_TOKEN")
	vaultCmd.PersistentFlags().StringVarP(&vaultPath, "path", "P", "", "Vault secret Path to Read/Write")
	vaultCmd.PersistentFlags().BoolVarP(&logical, "logical", "L", false, "Use Logical Api, default is KV2")
	vaultCmd.PersistentFlags().StringVarP(&kvMount, "mount", "M", kvMount, "Mount Path of the Secret engine")
	_ = vaultCmd.MarkFlagRequired("path")
	vaultReadCmd.Flags().BoolVarP(&jsonOut, "json", "J", false, "output as json")
	vaultWriteCmd.Flags().String("data_file", "", "Path to the json encoded file with the data to read from")
	vaultCmd.AddCommand(vaultReadCmd)
	vaultCmd.AddCommand(vaultWriteCmd)
	vaultCmd.AddCommand(vaultListCmd)
}

func vaultRead(_ *cobra.Command, args []string) error {
	var (
		err error
		key string
		vc  *vault.Client
		vs  *vault.Secret
		kvs *vault.KVSecret
	)
	log.Debugf("Vault Read entered for path '%s'", vaultPath)
	vc, _ = pwlib.VaultConfig(vaultAddr, vaultToken)
	if len(args) > 0 {
		key = args[0]
		log.Debugf("Selected Key '%s'", key)
	}
	if logical {
		vs, err = pwlib.VaultRead(vc, vaultPath)
		if err == nil {
			if vs != nil {
				log.Debug("Vault Read OK")
				err = printData(vs.Data, key)
			} else {
				err = fmt.Errorf("no entries returned")
			}
		}
	} else {
		kvs, err = pwlib.VaultKVRead(vc, kvMount, vaultPath)
		if err == nil {
			if kvs != nil {
				log.Debug("Vault KVRead OK")
				err = printData(kvs.Data, key)
			} else {
				err = fmt.Errorf("no entries returned")
			}
		}
	}
	if err == nil {
		log.Info("Vault Data successfully processed")
	}
	return err
}

func vaultWrite(cmd *cobra.Command, args []string) error {
	var (
		err       error
		datafile  string
		content   []byte
		vaultData map[string]interface{}
		vc        *vault.Client
	)

	log.Debug("Vault Write entered")
	if len(args) > 0 {
		content = []byte(args[0])
	} else {
		datafile, err = cmd.Flags().GetString("data_file")
		if err == nil {
			//nolint gosec
			content, err = os.ReadFile(datafile)
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
	err = json.Unmarshal(content, &vaultData)
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
		err       error
		vc        *vault.Client
		vs        *vault.Secret
		vaultkeys []interface{}
	)

	log.Debugf("Vault list entered for path '%s'", vaultPath)
	vc, _ = pwlib.VaultConfig(vaultAddr, vaultToken)
	vp := vaultPath
	if vp == "" {
		vp = "/"
	}
	if !logical {
		vp = path.Join(kvMount, "metadata", vp)
		log.Debugf("expand kv path for api to %s", vp)
	}
	vs, err = pwlib.VaultList(vc, vp)
	if err == nil {
		if vs != nil {
			vaultkeys = vs.Data["keys"].([]interface{})
			l := len(vaultkeys)
			log.Infof("Vault List returned %d entries", l)
			for _, k := range vaultkeys {
				fmt.Println(k)
			}
		} else {
			err = fmt.Errorf("no Entries returned")
		}
	} else {
		err = fmt.Errorf("list command failed:%s", err)
	}
	return err
}

func printData(vaultData map[string]interface{}, key string) (err error) {
	var jsonData []byte
	sysKey := strings.ReplaceAll(vaultPath, ":", "_")
	if key != "" {
		value, ok := vaultData[key].(string)
		if ok {
			fmt.Printf("%s", value)
		} else {
			err = fmt.Errorf("key '%s' not found", key)
		}
	} else {
		if len(vaultData) > 0 {
			if jsonOut {
				jsonData, err = json.Marshal(vaultData)
				if err != nil {
					err = fmt.Errorf("cannot generate json output:%s", err)
					return
				}
				fmt.Printf("%s\n", jsonData)
			} else {
				for k, v := range vaultData {
					fmt.Printf("%s:%s:%v\n", sysKey, k, v)
				}
			}
		} else {
			err = fmt.Errorf("no data found")
		}
	}
	return
}
