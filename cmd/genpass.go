// Package cmd Commands
package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

const defaultPasswordProfilesetFilename = "password_profiles.yaml"

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
	newCmd.Flags().StringP("special_chars", "s", "", "define allowed special chars")
	newCmd.Flags().StringP("profile", "p", "", "set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'")
	newCmd.Flags().StringP("profileset", "P", "", "set profile to existing named profile set")
	newCmd.Flags().String("password_profiles", "", "filename for loading password profiled")
	newCmd.Flags().BoolP("list_profiles", "l", false, "list existing profiles only")
	RootCmd.AddCommand(newCmd)
}

func genpass(cmd *cobra.Command, _ []string) error {
	log.Debugf("generate password called")
	var err error
	var pps pwlib.PasswordProfileSet
	data := ""
	l, _ := cmd.Flags().GetBool("list_profiles")
	if l {
		data, err = listProfiles(cmd)
	} else {
		pps, err = getPasswordProfileSet(cmd)
		if err != nil {
			return err
		}
		_, _ = pps.Load()
		data, err = pwlib.GenPasswordProfile(pps)
	}

	if err == nil {
		fmt.Println(data)
		return nil
	}
	return err
}

func listProfiles(cmd *cobra.Command) (string, error) {
	fn, _ := cmd.Flags().GetString("password_profiles")
	passwordProfiles, e := loadPasswordProfileSets(fn)
	if e != nil {
		return "", e
	}
	d, e := yaml.Marshal(passwordProfiles)
	if e != nil {
		e = fmt.Errorf("cannot marshal profile yaml to string:%v", e)
		return "", e
	}
	data := string(d)
	data = fmt.Sprintf("---\r\n# Default profile : %s \r\n%s", defaultProfileSetName, data)
	log.Debugf("list profilesets returned\r\n%s", data)
	return data, e
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
		s = defaultProfileSetName
		log.Infof("assume default profileset %s", s)
	}
	var pp pwlib.PasswordProfile
	var passwordProfiles pwlib.PasswordProfileSets
	if s != "" {
		success := false
		log.Debugf("got parameter profileset=%s", s)
		passwordProfiles, err = loadPasswordProfileSets(fn)
		if err != nil {
			err = fmt.Errorf("error loading password profiles: %s", err)
			return
		}
		pps, success = passwordProfiles[s]
		if !success {
			err = fmt.Errorf("profileset %s not found", s)
			return
		}
		_, _ = pps.Load()
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

func loadPasswordProfileSets(fn string) (passwordProfiles pwlib.PasswordProfileSets, err error) {
	var vps pwlib.PasswordProfileSets
	passwordProfiles, err = loadDefaultPasswordProfiles()
	if err != nil {
		return
	}

	fn = determineProfileFilename(fn)
	vps, err = loadExternalPasswordProfiles(fn)
	if err != nil {
		return
	}
	passwordProfiles, err = mergePasswordProfiles(passwordProfiles, vps, fn)
	return
}

func loadDefaultPasswordProfiles() (pwlib.PasswordProfileSets, error) {
	pps, err := pwlib.LoadPasswordProfileSets(defaultProfileSets)
	if err != nil {
		return nil, fmt.Errorf("error loading default password profiles: %s", err)
	}
	return pps, nil
}

func determineProfileFilename(fn string) string {
	if fn == "" {
		fn = viper.GetString("password_profiles")
	}
	if fn == "" {
		fn = defaultPasswordProfilesetFilename
	}
	return fn
}

func loadExternalPasswordProfiles(fn string) (vps pwlib.PasswordProfileSets, err error) {
	searchPaths := []string{viper.ConfigFileUsed(), ".", path.Join(home, ".pwcli"), path.Join(home, "etc"), "/etc/pwcli"}
	fn = common.FindFileInPath(fn, searchPaths)

	if len(fn) == 0 {
		log.Debugf("Password profile file not found in %s", strings.Join(searchPaths, ", "))
		return nil, nil
	}
	content := ""
	log.Debugf("try to load password profiles from file '%s'", fn)
	content, err = common.ReadFileToString(fn)
	if err != nil {
		return nil, fmt.Errorf("error reading password profiles from '%s': %s", fn, err)
	}
	if len(content) > 0 {
		log.Debugf("loading password profiles from '%s'", fn)
		vps, err = pwlib.LoadPasswordProfileSets(content)
		if err != nil {
			return nil, fmt.Errorf("error loading password profiles from '%s': %s", fn, err)
		}
		log.Debugf("loaded %d password profiles from '%s'", len(vps), fn)
	}
	return vps, nil
}

func mergePasswordProfiles(defaultProfiles, externalProfiles pwlib.PasswordProfileSets, fn string) (pwlib.PasswordProfileSets, error) {
	if len(externalProfiles) > 0 {
		ppf, err := common.MergeMaps(defaultProfiles, externalProfiles)
		if err != nil || ppf == nil {
			return nil, fmt.Errorf("error merging password profiles from '%s': %v", fn, err)
		}
		return ppf.(pwlib.PasswordProfileSets), nil
	}
	return defaultProfiles, nil
}
