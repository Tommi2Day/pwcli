// Package cmd Commands
package cmd

import (
	"fmt"

	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// nolint gosec
const defaultPasswordProfiles = `
default:
  profile:
    length: 16
    upper: 1
    lower: 1
    digits: 1
    specials: 1
    first_is_char: true
  special_chars: "!#=@&()"
easy:
  profile:
    length: 8
    upper: 1
    lower: 1
    digits: 1
    specials: 0
strong:
  profile:
    length: 32
    upper: 2
    lower: 2
    digits: 2
    specials: 2
    first_is_char: true
  special_chars: "!#=@&()"
`

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:          "genpass",
	Aliases:      []string{"gen", "new"},
	Short:        "generate new password for the given profile",
	Long:         `this will generate a random password according the given profile`,
	RunE:         genpass,
	SilenceUsage: true,
}
var passwordProfiles pwlib.PasswordProfileSets

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
	newCmd.Flags().StringP("special_chars", "s", "", "define allowed special chars")
	newCmd.Flags().StringP("profile", "p", "", "set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'")
	newCmd.Flags().StringP("profileset", "P", "", "set profile to existing named profile set")
	newCmd.Flags().String("password_profiles", "", "filename for loading password profiled")
	RootCmd.AddCommand(newCmd)
}

func genpass(cmd *cobra.Command, _ []string) error {
	log.Debugf("generate password called")
	pps, err := getPasswordProfileSet(cmd)
	if err != nil {
		return err
	}
	_, _ = pps.Load()
	password, err := pwlib.GenPasswordProfile(pps)
	if err == nil {
		fmt.Println(password)
		return nil
	}
	return err
}

func getPasswordProfileSet(cmd *cobra.Command) (pps pwlib.PasswordProfileSet, err error) {
	s, _ := cmd.Flags().GetString("profileset")
	ch, _ := cmd.Flags().GetString("special_chars")
	p, _ := cmd.Flags().GetString("profile")
	fn, _ := cmd.Flags().GetString("password_profiles")
	if s != "" && p != "" {
		err = fmt.Errorf("profileset and profile are mutually exclusive")
		return
	}
	if s == "" && p == "" {
		s = "default"
		log.Info("use default profileset")
	}
	var pp pwlib.PasswordProfile

	if s != "" {
		pps, err = preparePasswordProfileSet(s, fn)
		if err != nil {
			return
		}
	} else {
		log.Debugf("got parameter profile=%s", p)
		pp, err = pwlib.GetPasswordProfileFromString(p)
		if err != nil {
			err = fmt.Errorf("error setting password profile '%s': %s", p, err)
			return
		}
		pps = pwlib.PasswordProfileSet{Profile: pp}
	}
	if len(ch) > 0 {
		pps.SpecialChars = ch
	}
	return
}

func preparePasswordProfileSet(s string, fn string) (pps pwlib.PasswordProfileSet, err error) {
	log.Debugf("got parameter profileset=%s", s)
	// load default password profiles
	passwordProfiles, err = pwlib.LoadPasswordProfileSets(defaultPasswordProfiles)
	if err != nil {
		err = fmt.Errorf("error loading default password profiles: %s", err)
		return
	}

	// load external password profiles from file
	if fn == "" {
		fn = viper.GetString("password_profiles")
	}

	if fn != "" {
		var vps pwlib.PasswordProfileSets
		searchPaths := []string{".", "~", "~/", "~/.pwcli", "~/.pwcli/", "/etc/pwcli", viper.ConfigFileUsed()}
		fn = common.FindFileInPath(fn, searchPaths)
		log.Debugf("load password profiles from file '%s'", fn)
		cps, e := common.ReadFileToString(fn)
		if e != nil {
			err = fmt.Errorf("error reading password profiles from '%s': %s", fn, e)
			return
		}
		if len(cps) > 0 {
			vps, e = pwlib.LoadPasswordProfileSets(cps)
			if e != nil {
				err = fmt.Errorf("error loading password profiles from '%s': %s", fn, e)
				return
			}
			log.Debugf("loaded %d password profiles from '%s'", len(vps), fn)
			// merge default and loaded password sets
			ppf, e1 := common.MergeMaps(passwordProfiles, vps)
			if e1 != nil || ppf == nil {
				err = fmt.Errorf("error merging password profiles from '%s': %v", fn, e1)
				return
			}
			passwordProfiles = ppf.(pwlib.PasswordProfileSets)
		}
	}
	pps, success := passwordProfiles[s]
	if !success {
		err = fmt.Errorf("profileset %s not found", s)
		return
	}
	return
}
