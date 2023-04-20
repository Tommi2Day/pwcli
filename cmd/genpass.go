// Package cmd Commands
package cmd

import (
	"fmt"

	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

const defaultProfile = "10 1 1 1 0 1"
const defaultSpecials = pwlib.SpecialChar

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:          "genpass",
	Aliases:      []string{"gen", "new"},
	Short:        "generate new password for the given profile",
	Long:         `this will generate a random password according the given profile`,
	RunE:         genpass,
	SilenceUsage: true,
}

func genpass(cmd *cobra.Command, _ []string) error {
	log.Debugf("generate password called")
	s, _ := cmd.Flags().GetString("special_chars")
	pwlib.SetSpecialChars(s)
	p, _ := cmd.Flags().GetString("profile")
	profile, err := setPasswordProfile(p)
	if err != nil {
		return err
	}
	password, err := pwlib.GenPassword(profile.Length, profile.Upper, profile.Lower, profile.Digits, profile.Special, profile.Firstchar)
	if err == nil {
		fmt.Println(password)
		return nil
	}
	return err
}

func init() {
	// hide unused flags
	newCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		// Hide flag for this command
		_ = command.Flags().MarkHidden("app")
		_ = command.Flags().MarkHidden("keydir")
		_ = command.Flags().MarkHidden("datadir")
		_ = command.Flags().MarkHidden("config")
		_ = command.Flags().MarkHidden("method")
		// Call parent help func
		command.Parent().HelpFunc()(command, strings)
	})
	newCmd.Flags().StringP("special_chars", "s", defaultSpecials, "define allowed special chars")
	newCmd.Flags().StringP("profile", "p", defaultProfile, "set profile string as numbers of 'length Upper Lower Digits Special FirstcharFlag(0/1)'")
	RootCmd.AddCommand(newCmd)
}
