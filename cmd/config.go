// Package cmd Commands
package cmd

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommi2day/gomodules/common"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "handle config settings",
	Long:  `Allows read and write application config`,
}

var printCfgCmd = &cobra.Command{
	Use:          "print",
	Aliases:      []string{"list", "show"},
	Short:        "print current config in json format",
	RunE:         printConfig,
	SilenceUsage: true,
}

var saveCfgCmd = &cobra.Command{
	Use:          "save",
	Short:        "save current config parameter to file",
	RunE:         saveConfig,
	SilenceUsage: true,
}

var getCfgCmd = &cobra.Command{
	Use:          "get",
	Short:        "return value for key of running config",
	Long:         `return value for key of running config, pass the viper key as argument or using -k flag`,
	RunE:         getConfig,
	SilenceUsage: true,
}

func printConfig(_ *cobra.Command, _ []string) error {
	log.Debug("print config called")
	m := viper.AllSettings()
	c, err := common.StructToJSON(m)
	if err != nil {
		err = fmt.Errorf("error loading config: %s", err)
		return err
	}
	log.Debugf("config print:\n%s", c)
	fmt.Println(c)
	return nil
}
func saveConfig(cmd *cobra.Command, _ []string) error {
	var err error
	log.Debug(" Save config entered")
	force, _ := cmd.Flags().GetBool("force")
	filename, _ := cmd.Flags().GetString("filename")
	// try to get app from config
	if app == "" {
		app = viper.GetString("app")
	}
	if app == "" {
		app = configName
		viper.Set("app", app)
	}
	if filename == "" {
		filename = app + "." + configType
	}
	cfDir := path.Dir(filename)
	if !common.IsDir(cfDir) {
		err = os.MkdirAll(cfDir, 0700)
		if err != nil {
			log.Errorf("failed to create config directory %s: %s, choose another config file using --config", cfDir, err)
			return err
		}
		log.Infof("created config directory %s", cfDir)
	}
	viper.SetConfigFile(filename)
	log.Debugf("use filename '%s' to save", filename)
	if force {
		_, err = os.Stat(filename)
		if err == nil {
			log.Infof("Overwrite existing config")
		}
		err = viper.WriteConfigAs(filename)
	} else {
		err = viper.SafeWriteConfigAs(filename)
	}
	if err != nil {
		log.Errorf("Save config Error: %s", err)
	}
	log.Infof("config saved to '%s'", filename)
	fmt.Println("DONE")
	return err
}

func getConfig(cmd *cobra.Command, argv []string) error {
	var err error
	err = nil
	log.Debug(" get config entered")
	key, _ := cmd.Flags().GetString("key")
	log.Debugf("key as flag '%s', args: %d ", key, len(argv))
	if len(argv) > 0 {
		key = argv[0]
		log.Debugf("got key as arg '%s' ", key)
	}
	if key == "" {
		err = fmt.Errorf("need key to get")
		return err
	}

	v := viper.Get(key)
	if v == nil {
		log.Debugf("no value found for key %s", key)
		err = fmt.Errorf("no value found for key %s", key)
		return err
	}
	vs := fmt.Sprintf("%v", v)
	log.Infof("config value for key %s is %s", key, vs)
	fmt.Println(vs)

	return err
}
func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(printCfgCmd)
	configCmd.AddCommand(saveCfgCmd)
	configCmd.AddCommand(getCfgCmd)
	saveCfgCmd.Flags().StringP("filename", "f", "", "FileName to write")
	saveCfgCmd.Flags().Bool("force", false, "force overwrite")
	getCfgCmd.Flags().StringP("key", "k", "", "key to get")
}
