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
	noPromptFlag   = false

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
	typeKMS         = "kms"
	typePlain       = "plain"
	typeEnc         = "enc"
	typeAGE         = "age"
	typeGPG         = "gpg"
	typeGO          = "go"
	typeOpenSSL     = "openssl"
	typeGopass      = "gopass"
	defaultType     = "openssl"
)

// hideFlags hides the listed flags from a command's help output without
// affecting flag parsing.  It must not be called on a parent command that
// has subcommands (cobra will recurse and overflow the stack).
func hideFlags(command *cobra.Command, flags ...string) {
	original := command.HelpFunc()
	command.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		for _, f := range flags {
			_ = cmd.Flags().MarkHidden(f)
		}
		original(cmd, args)
	})
}

// hideGlobalFlags hides the key/password-file management global flags that
// are unrelated to the given command.  Pass extra names to hide additional
// flags (e.g. "no-prompt" for commands that never prompt).
func hideGlobalFlags(command *cobra.Command, extra ...string) {
	hideFlags(command, append([]string{"app", "keydir", "datadir", "config", "method"}, extra...)...)
}

func init() {
	cobra.OnInitialize(initConfig)
	initFlags()
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
		home, err = os.UserHomeDir()
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
	log.Debugf("Flags processed: debug=%v, info=%v, method=%s, app=%s", debugFlag, infoFlag, method, app)

	// sync flags with viper again just in case
	debugFlag = viper.GetBool("debug")
	infoFlag = viper.GetBool("info")
	method = viper.GetString("method")
	app = viper.GetString("app")
	keydir = viper.GetString("keydir")
	datadir = viper.GetString("datadir")

	configureLogging()

	// debug config file
	if cerr != nil {
		log.Debugf("Error using config for %s: %s", configName, cerr)
	}
	if haveConfig {
		cf := viper.ConfigFileUsed()
		log.Debugf("found configfile '%s'", cf)
		viper.Set("config", cf)
	}

	validateMethod()
	app = viper.GetString("app")
	// set pwlib config
	log.Debugf("InitConfig DEBUG: method=%s, app=%s, datadir=%s, keydir=%s\n", method, app, datadir, keydir)
	pc = pwlib.NewConfig(app, datadir, keydir, keypass, method)
	log.Debugf("PWConfig:%v", pc)
	switch {
	case keypass != "":
		log.Debug("keypass source: config or PW_KEYPASS env")
	case method == typeGPG:
		log.Debugf("keypass source: not set; %s will use GPG_PASSPHRASE env or gpg-agent", method)
	case method == typeAGE:
		log.Debugf("keypass source: not set; %s will use AGE_PASSPHRASE env or prompt if needed", method)
	case method == typeOpenSSL || method == typeGO:
		log.Debugf("keypass source: app name '%s' as default for method %s", app, method)
	default:
		log.Debugf("keypass source: not applicable for method %s", method)
	}
}

func configureLogging() {
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
	// no log output before here possible
	if unitTestFlag {
		log.SetOutput(RootCmd.OutOrStdout())
	}
}

func validateMethod() {
	if method == "" {
		method = defaultType
		log.Debugf("use default method %s ", defaultType)
	}
	if !slices.Contains(pwlib.PCmethods, method) {
		fmt.Println("Invalid method:", method)
		os.Exit(1)
	}
	if method == typeVault || method == typeKMS || method == typePlain || method == typeEnc {
		keypass = ""
	}
}

func initFlags() {
	RootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "", false, "verbose debug output")
	RootCmd.PersistentFlags().BoolVarP(&infoFlag, "info", "", false, "reduced info output")
	RootCmd.PersistentFlags().BoolVarP(&unitTestFlag, "unit-test", "", false, "redirect output for unit tests")
	_ = RootCmd.PersistentFlags().MarkHidden("unit-test")
	RootCmd.PersistentFlags().BoolVarP(&noLogColorFlag, "no-color", "", false, "disable colored log output")
	RootCmd.PersistentFlags().BoolVarP(&noPromptFlag, "no-prompt", "", false, "disable interactive prompts; return an error instead (for batch use)")
	RootCmd.PersistentFlags().StringVarP(&app, "app", "a", "", "name of application")
	RootCmd.PersistentFlags().StringVarP(&keydir, "keydir", "K", "", "directory of keys")
	RootCmd.PersistentFlags().StringVarP(&datadir, "datadir", "D", "", "directory of password files")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file name")
	RootCmd.PersistentFlags().StringVarP(&method, "method", "m", defaultType, "encryption method (openssl|go|enc|plain|vault|kms|age|gpg|gopass)")
}

// processConfig reads in config file and ENV variables if set.
func processConfig() (haveConfig bool, err error) {
	haveConfig = false
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
		haveConfig = true
		viper.Set("config", cfgFile)
		a := viper.GetString("app")
		if len(a) > 0 && app == "" {
			app = a
		}
	} else if cfgFile != "" {
		err = fmt.Errorf("%s:%v", cfgFile, err)
		cfgFile = ""
	}
	viper.Set("app", app)
	return
}

// promptKeypass interactively prompts for a key passphrase.
// Returns ("", nil) without prompting when --no-prompt is set.
func promptKeypass(label string) (string, error) {
	if noPromptFlag {
		return "", nil
	}
	return common.PromptPassword(label)
}

// methodUsesKeypass reports whether the encryption method decrypts a
// private key that may be passphrase-protected (age, gpg, openssl, go/rsa).
func methodUsesKeypass(m string) bool {
	switch m {
	case typeAGE, typeGPG, typeOpenSSL, typeGO:
		return true
	}
	return false
}

// checkKMSParams validates and applies the KMS key ID and endpoint for commands
// that use the kms method. No-op when method != typeKMS.
func checkKMSParams() error {
	if method != typeKMS {
		return nil
	}
	if kmsKeyID == "" {
		kmsKeyID = common.GetStringEnv("KMS_KEYID", "")
		log.Debugf("KMS KeyID from environment: '%s'", kmsKeyID)
	}
	if kmsKeyID == "" {
		return fmt.Errorf("need parameter kms_keyid to proceed")
	}
	if kmsEndpoint != "" {
		log.Debugf("use KMS endpoint %s", kmsEndpoint)
		_ = os.Setenv("KMS_ENDPOINT", kmsEndpoint)
	}
	log.Debugf("use KMS method with keyid %s", kmsKeyID)
	pc.KMSKeyID = kmsKeyID
	return nil
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
	if common.CmdFlagChanged(RootCmd, "app") {
		viper.Set("app", app)
	}
	if common.CmdFlagChanged(RootCmd, "keydir") {
		viper.Set("keydir", keydir)
	}
	if common.CmdFlagChanged(RootCmd, "datadir") {
		viper.Set("datadir", datadir)
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
	app = viper.GetString("app")
}
