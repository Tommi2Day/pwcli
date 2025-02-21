// Package cmd Commands
package cmd

import (
	"fmt"

	"github.com/spf13/viper"
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
	newCmd.Flags().StringP("profile", "p", defaultProfile, "set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'")
	newCmd.Flags().StringP("profileset", "P", "", "set profile to existing named profile set")
	// newCmd.MarkFlagsMutuallyExclusive("profileset", "profile")
	RootCmd.AddCommand(newCmd)
}

func genpass(cmd *cobra.Command, _ []string) error {
	log.Debugf("generate password called")
	s, _ := cmd.Flags().GetString("profileset")
	ch, _ := cmd.Flags().GetString("special_chars")
	p, _ := cmd.Flags().GetString("profile")
	if s != "" {
		log.Debugf("got parameter profileset=%s", s)
		ps := viper.GetStringMapString("password_profiles." + s)
		if len(ps) > 0 {
			if v, ok := ps["profile"]; ok {
				p = v
				log.Debugf("got parameter profile from parameterset %s", p)
			} else {
				log.Debugf("parameterset profile definition not found")
				return fmt.Errorf("parameterset profile definition not found")
			}
			if v, ok := ps["special_chars"]; ok {
				ch = v
				log.Debugf("got parameter special_chars from parameterset: %s", ch)
			} else {
				log.Debugf("parameter special_chars from parameterset not set, use default: %s", ch)
			}
		} else {
			log.Debugf("profilesset %s not found", s)
			return fmt.Errorf("profileset %s not found", s)
		}
	}
	pwlib.SetSpecialChars(ch)
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
