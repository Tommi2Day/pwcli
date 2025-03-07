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
	home           string
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
	// typeGopass      = "gopass"
	typeKMS   = "kms"
	typePlain = "plain"
	typeEnc   = "enc"
	// typeGPG         = "gpg"
	typeGO      = "go"
	typeOpenSSL = "openssl"
	defaultType = "openssl"
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
	RootCmd.PersistentFlags().StringVarP(&method, "method", "m", defaultType, "encryption method (openssl|go|gopass|enc|plain|vault|kms)")
	// don't have variables populated here
	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.Fatal(err)
	}
}

// Execute run application
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}
}

func initConfig() {
	var err error
	if cfgFile == "" {
		// Use config file from the flag.
		// Find home directory.
		home, err = homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in $HOME/etc $HOME/.pwcli, current directory and /etc/pwcli.
		viper.AddConfigPath(".")
		pwc := path.Join(home, ".pwcli")
		viper.AddConfigPath(pwc)
		etc := path.Join(home, "etc")
		viper.AddConfigPath(etc)
		viper.AddConfigPath("/etc/pwcli")
	}
	// env var overrides
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix(configEnvPrefix)
	// env var `LDAP_USERNAME` will be mapped to `ldap.username`
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in.
	haveConfig, cerr := processConfig()

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
	if cerr != nil {
		log.Debugf("Error using %s config: %s", configType, cerr)
	}
	if haveConfig {
		cf := viper.ConfigFileUsed()
		log.Debugf("found configfile '%s'", cf)
		viper.Set("config", cf)
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
	app = viper.GetString("app")
	// set pwlib config
	pc = pwlib.NewConfig(app, datadir, keydir, keypass, method)
}

// processConfig reads in config file and ENV variables if set.
func processConfig() (haveConfig bool, err error) {
	haveConfig = false
	viper.SetConfigType(configType)
	if cfgFile == "" {
		// try loading config file <app>.yaml
		if app != "" {
			viper.SetConfigName(app)
			err = viper.ReadInConfig()
		}

		if app == "" || err != nil {
			// try loading default configfile
			viper.SetConfigName(configName)
			err = viper.ReadInConfig()
		}
	} else {
		// try loading named config file
		viper.SetConfigFile(cfgFile)
		err = viper.ReadInConfig()
	}
	if err == nil {
		log.Debugf("using config file %s", cfgFile)
		haveConfig = true
		viper.Set("config", cfgFile)
		a := viper.GetString("app")
		if len(a) > 0 && app == "" {
			app = a
			viper.Set("app", app)
		}
	} else if cfgFile != "" {
		err = fmt.Errorf("%s:%v", cfgFile, err)
		cfgFile = ""
	}
	return
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
	if common.CmdFlagChanged(RootCmd, "method") {
		viper.Set("method", method)
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
