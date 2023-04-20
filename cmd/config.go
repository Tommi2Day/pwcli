// Package cmd Commands
package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "handle config settings",
	Long:  `Allows read and write application config`,
}

var printCmd = &cobra.Command{
	Use:     "print",
	Aliases: []string{"read"},
	Short:   "prints to stdout",
	Long:    `Allows read and write application config`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("print config called")
		for k, v := range viper.AllSettings() {
			fmt.Printf("%s=%v\n", k, v)
		}
	},
	SilenceUsage: true,
}

var saveCmd = &cobra.Command{
	Use:          "save",
	Short:        "save commandline parameter to file",
	Long:         `write application config`,
	RunE:         saveConfig,
	SilenceUsage: true,
}

func saveConfig(cmd *cobra.Command, _ []string) error {
	var err error
	log.Debug(" Save config entered")
	force, _ := cmd.Flags().GetBool("force")
	filename := cfgFile
	if filename == "" {
		filename = viper.ConfigFileUsed()
	}
	if filename == "" {
		err = fmt.Errorf("need a config filename, eg. --config")
		return err
	}
	log.Debugf("use filename '%s'", filename)
	if force {
		_, err = os.Stat(filename)
		if os.IsExist(err) {
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

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(printCmd)
	configCmd.AddCommand(saveCmd)
	saveCmd.Flags().BoolP("force", "f", false, "force overwrite")
}
