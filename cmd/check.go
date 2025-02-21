// Package cmd Commands
package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommi2day/gomodules/pwlib"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:          "check",
	Short:        "checks a password to given profile",
	Long:         `Checks a password for charset and length rules`,
	RunE:         checkPassword,
	SilenceUsage: true,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires password to test as argument")
		}
		return nil
	},
}

func init() {
	// hide unused flags
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
	checkCmd.Flags().StringP("special_chars", "s", defaultSpecials, "define allowed special chars")
	checkCmd.Flags().StringP("profile", "p", defaultProfile, "set profile string as numbers of 'length Upper Lower Digits Special FirstcharFlag(0/1)'")
	checkCmd.Flags().StringP("profileset", "P", "", "set profile to existing named profile set")
	// checkCmd.MarkFlagsMutuallyExclusive("profileset", "profile")
	RootCmd.AddCommand(checkCmd)
}

func checkPassword(cmd *cobra.Command, args []string) error {
	log.Debug("check password profile called")
	var profile pwlib.PasswordProfile
	var err error
	log.Debugf("Args:%v", args)
	password := args[0]
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
	profile, err = setPasswordProfile(p)
	if err != nil {
		return err
	}
	if pwlib.DoPasswordCheck(password, profile.Length, profile.Upper, profile.Lower, profile.Digits, profile.Special, profile.Firstchar, "") {
		fmt.Println("SUCCESS")
		log.Infof("Password '%s' matches the given profile", password)
		return nil
	}
	err = fmt.Errorf("password '%s' matches NOT the given profile", password)
	return err
}

func setPasswordProfile(p string) (profile pwlib.PasswordProfile, err error) {
	if len(p) == 0 {
		p = defaultProfile
		log.Infof("Choose default profile %s", defaultProfile)
	}
	custom := strings.Split(p, " ")
	if len(custom) < 6 {
		err = fmt.Errorf("profile string should have 6 space separated numbers <length> <upper chars> <lower chars> <digits> <special chars> <do firstchar check(0/1)>")
		return
	}
	profile.Length, err = strconv.Atoi(custom[0])
	if err == nil {
		profile.Upper, err = strconv.Atoi(custom[1])
	}
	if err == nil {
		profile.Lower, err = strconv.Atoi(custom[2])
	}
	if err == nil {
		profile.Digits, err = strconv.Atoi(custom[3])
	}
	if err == nil {
		profile.Special, err = strconv.Atoi(custom[4])
	}
	if err == nil {
		f := custom[5]
		profile.Firstchar = f == "1"
	}
	return
}
