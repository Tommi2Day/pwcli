// Package cmd commands
package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/tommi2day/gomodules/common"

	"golang.org/x/exp/slices"

	"github.com/tommi2day/gomodules/pwlib"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	keydir         string
	datadir        string
	app            string
	keypass        string
	cfgFile        string
	method         string
	pc             *pwlib.PassConfig
	debugFlag      = false
	infoFlag       = false
	noLogColorFlag = false
	unitTestFlag   = false

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
	typeGopass      = "gopass"
	typeKMS         = "kms"
	typePlain       = "plain"
	typeEnc         = "enc"
	typeGPG         = "gpg"
	typeGO          = "go"
	typeOpenSSL     = "openssl"
	defaultType     = "openssl"
)

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "", false, "verbose debug output")
	RootCmd.PersistentFlags().BoolVarP(&infoFlag, "info", "", false, "reduced info output")
	RootCmd.PersistentFlags().BoolVarP(&unitTestFlag, "unit-test", "", false, "redirect output for unit tests")
	RootCmd.PersistentFlags().BoolVarP(&noLogColorFlag, "no-color", "", false, "disable colored log output")
	RootCmd.PersistentFlags().StringVarP(&app, "app", "a", configName, "name of application")
	RootCmd.PersistentFlags().StringVarP(&keydir, "keydir", "K", "", "directory of keys")
	RootCmd.PersistentFlags().StringVarP(&datadir, "datadir", "D", "", "directory of password files")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file name")
	RootCmd.PersistentFlags().StringVarP(&method, "method", "m", defaultType, "encryption method (openssl|go|gopass|enc|plain|vault)")
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

		// Search config in $HOME/etc and current directory.
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
	haveConfig, err := processConfig()

	// check flags
	processFlags()

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
		DisableColors:   noLogColorFlag,
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	}
	log.SetFormatter(logFormatter)

	if unitTestFlag {
		log.SetOutput(RootCmd.OutOrStdout())
	}
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
	if method == typeVault || method == typeKMS || method == typePlain || method == typeEnc {
		keypass = ""
	}
	// set pwlib config
	pc = pwlib.NewConfig(app, datadir, keydir, keypass, method)
}

// processConfig reads in config file and ENV variables if set.
func processConfig() (bool, error) {
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
	}
	return haveConfig, err
}

func processFlags() {
	if common.CmdFlagChanged(RootCmd, "debug") {
		viper.Set("debug", debugFlag)
	}
	if common.CmdFlagChanged(RootCmd, "info") {
		viper.Set("info", infoFlag)
	}
	if common.CmdFlagChanged(RootCmd, "no-color") {
		viper.Set("no-color", noLogColorFlag)
	}
	if keydir == "" {
		keydir = viper.GetString("keydir")
	}
	if datadir == "" {
		datadir = viper.GetString("datadir")
	}
	keypass = viper.GetString("keypass")
	debugFlag = viper.GetBool("debug")
	infoFlag = viper.GetBool("info")
	method = viper.GetString("method")
}
