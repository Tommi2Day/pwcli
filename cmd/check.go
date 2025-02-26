// Package cmd Commands
package cmd

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/tommi2day/gomodules/pwlib"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:          "checkpass",
	Short:        "checks a password to given profile",
	Long:         `Checks a password for charset and length rules`,
	RunE:         checkPassword,
	Aliases:      []string{"check"},
	SilenceUsage: true,
	/*
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires password to test as argument")
			}
			return nil
		},
	*/
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
	checkCmd.Flags().StringP("special_chars", "s", "", "define allowed special chars")
	checkCmd.Flags().StringP("profile", "p", "", "set profile string as numbers of 'length Upper Lower Digits Special FirstcharFlag(0/1)'")
	checkCmd.Flags().StringP("profileset", "P", "", "set profile to existing named profile set")
	checkCmd.Flags().String("password_profiles", "", "filename for loading password profiled")
	checkCmd.Flags().BoolP("list_profiles", "l", false, "list existing profiles only")
	RootCmd.AddCommand(checkCmd)
}

func checkPassword(cmd *cobra.Command, args []string) error {
	log.Debug("check password profile called")
	var err error
	var pps pwlib.PasswordProfileSet
	data := ""
	l, _ := cmd.Flags().GetBool("list_profiles")
	if l {
		data, err = listProfiles(cmd)
		if err != nil {
			return err
		}
		fmt.Println(data)
		return nil
	}
	if len(args) > 0 {
		log.Debugf("Args:%v", args)
		password := args[0]
		pps, err = getPasswordProfileSet(cmd)

		if err != nil {
			return err
		}
		profile, cs := pps.Load()
		if pwlib.DoPasswordCheck(password, profile, cs) {
			fmt.Println("SUCCESS")
			log.Infof("Password '%s' matches the given profile", password)
			return nil
		}
		err = fmt.Errorf("password '%s' matches NOT the given profile", password)
		return err
	}
	log.Debug("No args given")
	err = errors.New("requires password to check as argument")
	return err
}
