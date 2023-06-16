// Package cmd commands
package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/tommi2day/gomodules/pwlib"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var keydir string
var datadir string
var app string
var keypass string
var cfgFile string
var method string
var pc *pwlib.PassConfig
var debugFlag = false
var infoFlag = false
var (
	// RootCmd entry point to start
	RootCmd = &cobra.Command{
		Use:           "pwcli",
		Short:         "pwcli – Password generation and encryption Tools",
		Long:          ``,
		SilenceErrors: true,
	}
)

const (
	// allows you to override any config values using
	// env APP_MY_VAR = "MY_VALUE"
	// e.g. export APP_LDAP_USERNAME test
	// maps to ldap.username
	configEnvPrefix = "PW"
	configName      = "pwcli"
	configType      = "yaml"
	typeVault       = "vault"
	defaultType     = "openssl"
)

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "", false, "verbose debug output")
	RootCmd.PersistentFlags().BoolVarP(&infoFlag, "info", "", false, "reduced info output")
	RootCmd.PersistentFlags().StringVarP(&app, "app", "a", configName, "name of application")
	RootCmd.PersistentFlags().StringVarP(&keydir, "keydir", "K", "", "directory of keys")
	RootCmd.PersistentFlags().StringVarP(&datadir, "datadir", "D", "", "directory of password files")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file name")
	RootCmd.PersistentFlags().StringVarP(&method, "method", "m", defaultType, "encryption method (openssl|go|enc|plain|vault)")
	// don't have variables populated here
	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.Fatal(err)
	}
}

// Execute run application
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetConfigType(configType)
	viper.SetConfigName(configName)
	if cfgFile == "" {
		// Use config file from the flag.
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home/etc and current directory).
		etc := path.Join(home, "etc")
		viper.AddConfigPath(etc)
		viper.AddConfigPath(".")
	} else {
		// set filename form cli
		viper.SetConfigFile(cfgFile)
	}

	// env var overrides
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix(configEnvPrefix)
	// env var `LDAP_USERNAME` will be mapped to `ldap.username`
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	haveConfig := false
	if err == nil {
		cfgFile = viper.ConfigFileUsed()
		haveConfig = true
		viper.Set("config", cfgFile)
		a := viper.GetString("app")
		if len(a) > 0 {
			app = a
		}
		if RootCmd.Flags().Lookup("debug").Changed {
			viper.Set("debug", debugFlag)
		}
		if RootCmd.Flags().Lookup("info").Changed {
			viper.Set("info", infoFlag)
		}
		keypass = viper.GetString("keypass")
		debugFlag = viper.GetBool("debug")
		infoFlag = viper.GetBool("info")
		method = viper.GetString("method")
		if keydir == "" {
			keydir = viper.GetString("keydir")
		}
		if datadir == "" {
			datadir = viper.GetString("datadir")
		}
	}

	// logger settings
	log.SetLevel(log.ErrorLevel)
	switch {
	case debugFlag:
		// report function name
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	case infoFlag:
		log.SetLevel(log.InfoLevel)
	}

	logFormatter := &prefixed.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	}
	log.SetFormatter(logFormatter)

	// debug config file
	if haveConfig {
		log.Debugf("found configfile %s", cfgFile)
	} else {
		log.Debugf("Error using %s config: %s", configType, err)
	}

	// validate method
	if method == "" {
		method = defaultType
		log.Debugf("use default method %s ", defaultType)
	}
	if !slices.Contains(pwlib.Methods, method) {
		fmt.Println("Invalid method:", method)
		os.Exit(1)
	}
	if method == typeVault {
		keypass = ""
	}
	// set pwlib config
	pc = pwlib.NewConfig(app, datadir, keydir, keypass, method)
}
