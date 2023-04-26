package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/pwlib"
)

// totpCmd represents the totp command
var totpCmd = &cobra.Command{
	Use:          "totp",
	Short:        "generate totp code from secret",
	Long:         `generate a standard 6 digit auth/mfa code for given secret with --secret or TOTP_SECRET env`,
	SilenceUsage: true,
	RunE:         genTOTP,
}

func genTOTP(cmd *cobra.Command, _ []string) error {
	log.Debug("TOTP called")
	secret, _ := cmd.Flags().GetString("secret")
	if secret == "" {
		secret = os.Getenv("TOTP_SECRET")
		log.Debugf("use secret from env: %s", secret)
	}
	if secret == "" {
		err := fmt.Errorf("no secret given, use --secret or Env TOTP_SECRET")
		return err
	}
	totp, err := pwlib.GetOtp(secret)
	if err == nil {
		log.Infof("TOTP returned %s", totp)
		fmt.Println(totp)
	} else {
		err = fmt.Errorf("TOTP generation failed:%s", err)
	}
	return err
}

func init() {
	checkCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		// Hide flag for this command
		_ = command.Flags().MarkHidden("app")
		_ = command.Flags().MarkHidden("keydir")
		_ = command.Flags().MarkHidden("datadir")
		_ = command.Flags().MarkHidden("config")
		_ = command.Flags().MarkHidden("method")
		// Call parent help func
		command.Parent().HelpFunc()(command, strings)
	})
	RootCmd.AddCommand(totpCmd)
	// don't have variables populated here
	totpCmd.Flags().StringP("secret", "s", "", "totp secret to generate code from")
}
