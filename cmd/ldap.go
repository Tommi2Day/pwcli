package cmd

import (
	"fmt"
	"os"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/ldaplib"
	"github.com/tommi2day/gomodules/pwlib"
)

var ldapServer = ""
var ldapBindDN = ""
var ldapBindPassword = ""
var ldapBaseDN = ""
var targetDN = ""
var ldapPort = 0
var ldapInsecure = false
var ldapTLS = false
var ldapTimeout = 20
var ldapGroupBase = ""
var ldapTargetUser = ""
var inputReader = os.Stdin

const ldapPublicKeyObjectClass = "ldapPublicKey"
const ldapSSHAttr = "sshPublicKey"

// nolint gosec
const ldapPasswordProfile = "easy"

// var ldapUserContext = "ou=users"

var ldapCmd = &cobra.Command{
	Use:   "ldap",
	Short: "commands related to ldap",
}

// ldapPassCmd represents the new command
var ldapPassCmd = &cobra.Command{
	Use:     "setpass",
	Aliases: []string{"change-password"},
	Short:   "change LDAP Password for given User per DN",
	Long: `set new ldap password by --new-password or Env LDAP_NEW_PASSWORD for the actual bind DN or as admin bind for a target DN.
if no new password given some systems will generate a password`,
	RunE:         setLdapPass,
	SilenceUsage: true,
}

// ldapPassCmd represents the new command
var ldapSSHCmd = &cobra.Command{
	Use:          "setssh",
	Aliases:      []string{"change-sshpubkey"},
	Short:        "Set public SSH Key to LDAP DN",
	Long:         `set new ssh public key(attribute sshPublicKey) for a given User per DN, the key must be in a file given by --sshpubkeyfile or default id_rsa.pub.`,
	RunE:         setSSHKey,
	SilenceUsage: true,
}

// ldapShowCmd represents the new command
var ldapShowCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"show-attributes", "attributes"},
	Short:   "Show attributes of LDAP DN",
	Long: `This command shows the attributes off the own User(Bind User) or
you may lookup a User cn and show the attributes of the first entry returned.`,
	RunE:         showAttributes,
	SilenceUsage: true,
}

// ldapGroupCmd represents the new command
var ldapGroupCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"show-groups", "group-membership"},
	Short:   "Show the group memberships of the given DN",
	Long: `This command shows the group membership of  own User(Bind User) or
you may lookup a User cn and if found show the groups of the first entry returned`,
	RunE:         showGroups,
	SilenceUsage: true,
}

// ldapMemberCmd represents the new command
var ldapMemberCmd = &cobra.Command{
	Use:     "members",
	Aliases: []string{"show-members", "group-members"},
	Short:   "Search the members  of the given group CN",
	Long: `This command shows members of a given group or
you may lookup a group cn and if found show the members of the groups`,
	RunE:         showMembers,
	SilenceUsage: true,
}
var hideFlags = func(command *cobra.Command, strings []string) {
	// Hide flag for this command
	_ = command.Flags().MarkHidden("app")
	_ = command.Flags().MarkHidden("keydir")
	_ = command.Flags().MarkHidden("datadir")
	_ = command.Flags().MarkHidden("config")
	_ = command.Flags().MarkHidden("method")
	// Call parent help func
	command.Parent().HelpFunc()(command, strings)
}

func showGroups(_ *cobra.Command, _ []string) error {
	log.Debugf("ldap groups called")
	lc, err := ldapLogin()
	if err != nil {
		log.Warnf("ldap login returned error %v", err)
		return err
	}

	// validate parameter
	if ldapGroupBase == "" {
		ldapGroupBase = ldapBaseDN
	}

	// lookup target user if given
	udn := ""
	if ldapTargetUser != "" {
		udn, err = lookupTargetUser(lc, ldapTargetUser)
		if err != nil {
			log.Errorf("%v", err)
			return err
		}
		if udn != "" {
			targetDN = udn
		}
	}
	log.Debugf("targetDN:%s", targetDN)
	// search for targetDN entry
	filter := fmt.Sprintf("(|(&(objectclass=groupOfUniqueNames)(uniqueMember=%s))(&(objectclass=groupOfNames)(member=%s)))", targetDN, targetDN)
	log.Debugf("ldap search for groups with filter %s", filter)
	entries, err := lc.Search(ldapGroupBase, filter, []string{"DN"}, ldap.ScopeWholeSubtree, ldap.DerefInSearching)
	if err != nil {
		log.Errorf("search for %s returned error %v", targetDN, err)
		return fmt.Errorf("search for %s returned error %v", targetDN, err)
	}
	if len(entries) == 0 {
		log.Warnf("no groups for %s found", targetDN)
		fmt.Printf("no groups for %s found", targetDN)
		return nil
	}
	fmt.Printf("DN '%s' is member of the following groups:\n", targetDN)
	for _, e := range entries {
		log.Infof("Group: %s", e.DN)
		fmt.Printf("Group: %s\n", e.DN)
	}
	return nil
}
func showMembers(cmd *cobra.Command, _ []string) error {
	log.Debugf("ldap members called")
	lc, err := ldapLogin()
	if err != nil {
		log.Warnf("ldap login returned error %v", err)
		return err
	}

	groupName, err := getGroupName(cmd)
	if err != nil {
		return err
	}

	if ldapGroupBase == "" {
		ldapGroupBase = ldapBaseDN
	}

	entries, err := searchGroupEntries(lc, ldapGroupBase, groupName)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		log.Warnf("no members for %s found", groupName)
		fmt.Printf("no members for %s found", groupName)
		return nil
	}

	displayGroupMembers(entries, groupName)
	return nil
}

func getGroupName(cmd *cobra.Command) (string, error) {
	g, err := cmd.Flags().GetString("group")
	if g == "" || err != nil {
		return "", fmt.Errorf("requires group name part in 'group' set")
	}
	return strings.ToLower(g), nil
}

func searchGroupEntries(lc *ldaplib.LdapConfigType, baseDN, groupName string) ([]*ldap.Entry, error) {
	filter := fmt.Sprintf("(cn=*%s*)", groupName)
	log.Debugf("ldap search for groups with filter %s", filter)
	entries, err := lc.Search(baseDN, filter, []string{"DN", "uniqueMember", "memberof", "member"}, ldap.ScopeWholeSubtree, ldap.DerefInSearching)
	if err != nil {
		log.Errorf("search for %s returned error %v", groupName, err)
		return nil, fmt.Errorf("search for %s returned error %v", groupName, err)
	}
	return entries, nil
}

func displayGroupMembers(entries []*ldap.Entry, groupName string) {
	fmt.Printf("search for '%s' returned members in following groups:\n", groupName)
	for _, entry := range entries {
		log.Infof("Group: %s", entry.DN)
		fmt.Printf("Group: %s\n", entry.DN)
		displayMembers(entry.Attributes)
		fmt.Println()
	}
}

func displayMembers(attributes []*ldap.EntryAttribute) {
	for _, attr := range attributes {
		if attr.Name == "uniqueMember" || attr.Name == "member" || attr.Name == "memberof" {
			for _, value := range attr.Values {
				log.Infof("    Member: %s", value)
				fmt.Printf("    Member: %s\n", value)
			}
		}
	}
}

func showAttributes(cmd *cobra.Command, _ []string) error {
	log.Debugf("ldap show called")
	lc, err := ldapLogin()
	if err != nil {
		log.Warnf("ldap login returned error %v", err)
		return err
	}

	// validate parameter
	attributes, _ := cmd.Flags().GetString("attributes")
	if attributes == "" {
		attributes = "*"
	}

	// lookup target user if given
	udn := ""
	if ldapTargetUser != "" {
		udn, err = lookupTargetUser(lc, ldapTargetUser)
		if err != nil {
			log.Errorf("%v", err)
			return err
		}
		if udn != "" {
			targetDN = udn
		}
	}
	log.Debugf("targetDN:%s", targetDN)

	// search for targetDN entry
	log.Debugf("ldap search for %s", targetDN)
	e, err := lc.RetrieveEntry(targetDN, "", attributes)
	if err != nil {
		log.Errorf("search for %s returned error %v", targetDN, err)
		return fmt.Errorf("search for %s returned error %v", targetDN, err)
	}
	if e == nil {
		log.Errorf("ldap search for %s returned no entry", targetDN)
		return fmt.Errorf("ldap search for %s returned no entry", targetDN)
	}
	fmt.Printf("DN '%s' has following attributes:\n", targetDN)
	values := e.Attributes
	for _, v := range values {
		name := v.Name
		for _, val := range v.Values {
			log.Infof("%s: %s", name, val)
			fmt.Printf("%s: %s\n", name, val)
		}
	}
	return nil
}

func promptPassword(label string) (pw string, err error) {
	prompt := promptui.Prompt{
		Label: label,
		Mask:  '*',
		Stdin: inputReader,
	}

	result, err := prompt.Run()
	if err != nil {
		return
	}
	pw = result
	return
}

func enterNewPassword() (pw string, err error) {
	pw, err = promptPassword("Enter NEW password")
	if err != nil {
		err = fmt.Errorf("error reading password: %v", err)
		return
	}
	log.Debugf("PW1: '%s'", pw)
	pw2 := ""
	if !unitTestFlag {
		pw2, err = promptPassword("Repeat NEW password")
	} else {
		// cannot use second promptui in unit tests
		pw2 = pw
	}
	log.Debugf("PW2: '%s'", pw)
	if err != nil {
		err = fmt.Errorf("error reading password: %v", err)
		return
	}
	if pw != pw2 {
		err = fmt.Errorf("passwords do not match")
	}
	return
}

func getNewLapPassword(cmd *cobra.Command) (newPassword string, err error) {
	generate, _ := cmd.Flags().GetBool("generate")
	if generate {
		log.Debugf("generated Ldap Password")
		pps, e := getPasswordProfileSet(cmd)
		if e != nil {
			log.Errorf("password profile set returned error %v", e)
			err = e
			return
		}
		log.Debugf("generated Password: %s", pps)
		newPassword, err = pwlib.GenPasswordProfile(pps)
		if err != nil {
			return
		}
		log.Infof("generated Password: %s", newPassword)
		fmt.Printf("generated Password: %s\n", newPassword)
	} else {
		newPassword, _ = cmd.Flags().GetString("new-password")
		if newPassword == "" {
			newPassword = os.Getenv("LDAP_NEW_PASSWORD")
			if newPassword != "" {
				log.Debugf("use new password from env: %s", newPassword)
			}
		}
		if newPassword == "" {
			fmt.Printf("Change password for %s\n", targetDN)
			newPassword, err = enterNewPassword()
			if err != nil {
				return
			}
		}
	}
	return
}
func setLdapPass(cmd *cobra.Command, _ []string) error {
	log.Debugf("ldap password called")
	// login to server
	lc, err := ldapLogin()
	if err != nil {
		log.Errorf("ldap login returned error %v", err)
		return err
	}
	// lookup target user if given
	udn := ""
	if ldapTargetUser != "" {
		udn, err = lookupTargetUser(lc, ldapTargetUser)
		if err != nil {
			log.Errorf("%v", err)
			return err
		}
		if udn != "" {
			targetDN = udn
		}
	}
	log.Debugf("targetDN: %s", targetDN)

	// validate parameter
	newPassword := ""
	newPassword, err = getNewLapPassword(cmd)
	if err != nil {
		return err
	}
	if newPassword == "" {
		// err = fmt.Errorf("no new password given, use --new_password or Env LDAP_NEW_PASSWORD")
		// return err
		log.Infof("even no new password given, it will be generated in some systems ldap system such as openldap")
	}

	//  for self write old password must be given und targetDN empty. for admin write targetDN must be given and old password empty
	oldPass := ""
	dn := targetDN
	if targetDN == ldapBindDN {
		oldPass = ldapBindPassword
		dn = ""
		log.Debugf("change password for myself")
	}
	// change password
	genPass := ""
	genPass, err = lc.SetPassword(dn, oldPass, newPassword)
	if err != nil {
		log.Errorf("ldap password change for %s returned error %v", targetDN, err)
		return fmt.Errorf("ldap password change for %s returned error %v", targetDN, err)
	}
	log.Infof("Password for %s changed", targetDN)
	if genPass == "" {
		genPass = newPassword
	} else {
		log.Infof("generated Password: %s", genPass)
		fmt.Printf("generated Password: %s\n", genPass)
	}
	l := lc.Conn
	_ = l.Close()

	// reconnect with new password to verify
	log.Debugf("reconnect with new password to verify")
	err = lc.Connect(targetDN, genPass)
	if err != nil {
		log.Errorf("ldap test bind to %s with new pass returned error %v", targetDN, err)
		return fmt.Errorf("ldap test bind to %s with new pass returned error %v", targetDN, err)
	}
	l = lc.Conn
	if l != nil {
		_ = l.Close()
	}
	log.Infof("SUCCESS: Password for %s changed and tested", targetDN)
	fmt.Printf("Password for %s changed and tested\n", targetDN)
	return nil
}

func setSSHKey(cmd *cobra.Command, _ []string) error {
	log.Debugf("ldap ssh key called")
	lc, err := ldapLogin()
	if err != nil {
		log.Warnf("ldap login returned error %v", err)
		return err
	}

	// lookup target user if given
	udn := ""
	if ldapTargetUser != "" {
		udn, err = lookupTargetUser(lc, ldapTargetUser)
		if err != nil {
			log.Errorf("%v", err)
			return err
		}
		if udn != "" {
			targetDN = udn
		}
	}
	log.Debugf("targetDN: %s", targetDN)
	// validate parameter
	sshPubKeyFile, _ := cmd.Flags().GetString("sshpubkeyfile")
	if sshPubKeyFile == "" {
		log.Warnf("sshpubkeyfile not given")
		return fmt.Errorf("sshpubkeyfile not given")
	}
	if !common.IsFile(sshPubKeyFile) {
		log.Warnf("sshpubkeyfile %s not found", sshPubKeyFile)
		return fmt.Errorf("sshpubkeyfile %s not found", sshPubKeyFile)
	}

	// read ssh key
	pubKey := ""
	pubKey, err = common.ReadFileToString(sshPubKeyFile)
	if err != nil {
		log.Warnf("sshpubkeyfile %s not readable", sshPubKeyFile)
		return fmt.Errorf("sshpubkeyfile %s not readable", sshPubKeyFile)
	}
	log.Debugf("got ssh key from file %s", sshPubKeyFile)
	// search for targetDN entry
	log.Debugf("ldap exact search for %s", targetDN)
	l := lc.Conn
	e, err := lc.RetrieveEntry(targetDN, "", "")
	if err != nil {
		log.Errorf("search for %s returned error %v", targetDN, err)
		return fmt.Errorf("search for %s returned error %v", targetDN, err)
	}
	if e == nil {
		log.Errorf("ldap search for %s returned no entry", targetDN)
		return fmt.Errorf("ldap search for %s returned no entry", targetDN)
	}
	log.Debugf("%s: look for objectclass %s ", targetDN, ldapSSHAttr)
	// check if attribute is assigned
	if !ldaplib.HasObjectClass(e, ldapPublicKeyObjectClass) {
		log.Errorf("objectclass %s not found for %s", ldapPublicKeyObjectClass, targetDN)
		return fmt.Errorf("objectclass %s not found for %s", ldapPublicKeyObjectClass, targetDN)
	}

	// check if attribute is already assigned or should added
	action := "replace"
	if !ldaplib.HasAttribute(e, ldapSSHAttr) {
		action = "add"
		log.Infof("attribute %s not found for %s, will be added", ldapSSHAttr, targetDN)
	}

	// change or add ssh key
	log.Debugf("change ssh key for %s", targetDN)
	err = lc.ModifyAttribute(targetDN, action, ldapSSHAttr, []string{pubKey})
	if err != nil {
		log.Errorf("ldap ssh key change for %s returned error %v", targetDN, err)
		return fmt.Errorf("ldap ssh key change for %s returned error %v", targetDN, err)
	}
	if err = verifySSHKey(lc, pubKey); err != nil {
		log.Errorf("%v", err)
		return err
	}
	log.Infof("SUCCESS: SSH Key for %s changed", targetDN)
	fmt.Printf("SSH Key for %s changed\n", targetDN)
	_ = l.Close()
	return nil
}

func verifySSHKey(lc *ldaplib.LdapConfigType, pubKey string) (err error) {
	var e *ldap.Entry
	// check if ssh key was changed
	log.Debugf("search for %s attribute %s to verify ssh key", targetDN, ldapSSHAttr)
	e, err = lc.RetrieveEntry(targetDN, "", ldapSSHAttr)
	if err != nil {
		log.Errorf("validate search for %s returned error %v", targetDN, err)
		return fmt.Errorf("validate search for %s returned error %v", targetDN, err)
	}
	actSSH := e.GetAttributeValue(ldapSSHAttr)
	if actSSH != pubKey {
		log.Errorf("ldap ssh key change for %s not successful, new value not as expected", targetDN)
		return fmt.Errorf("ldap ssh key change for %s not successful, new value not as expected", targetDN)
	}
	return
}
func lookupTargetUser(lc *ldaplib.LdapConfigType, user string) (dn string, err error) {
	log.Debugf("lookup user %s", user)
	if user == "" {
		return
	}
	if ldapBaseDN == "" {
		err = fmt.Errorf("no ldap base given")
		return
	}
	filter := fmt.Sprintf("(|(cn=%s)(uid=%s))", user, user)
	log.Debugf("search with filter %s from %s", user, ldapBaseDN)
	entries, err := lc.Search(ldapBaseDN, filter, []string{"DN"}, ldap.ScopeWholeSubtree, ldap.DerefInSearching)
	if err != nil {
		err = fmt.Errorf("ldap search for user %s returned error %v", user, err)
		return
	}
	l := len(entries)
	if l == 0 {
		err = fmt.Errorf("ldap search for user  %s returned no entry", user)
		return
	}
	if l > 1 {
		log.Debugf("search for user %s returned %d entries", user, l)
		list := make([]string, l)
		for i, e := range entries {
			list[i] = e.DN
		}
		prompt := promptui.Select{
			Label: "Select one of the following entries",
			Items: list,
		}

		_, dn, err = prompt.Run()

		if err != nil {
			fmt.Printf("select entry failed %v\n", err)
			return
		}
	} else {
		dn = entries[0].DN
	}
	log.Debugf("use dn %s for user %s", dn, user)
	return
}

func init() {
	ldapCmd.PersistentFlags().StringVarP(&ldapServer, "ldap.host", "H", "", "Hostname of Ldap Server")
	ldapCmd.PersistentFlags().IntVarP(&ldapPort, "ldap.port", "P", ldapPort, "ldap port to connect")
	ldapCmd.PersistentFlags().StringVarP(&ldapBaseDN, "ldap.base", "b", "", "Ldap Base DN ")
	ldapCmd.PersistentFlags().StringVarP(&targetDN, "ldap.targetdn", "T", "", "DN of target User for admin executed password change, empty for own entry (uses LDAP_BIND_DN)")
	ldapCmd.PersistentFlags().StringVarP(&ldapTargetUser, "ldap.targetuser", "U", "", "uid to search for targetDN")
	ldapCmd.PersistentFlags().StringVarP(&ldapBindDN, "ldap.binddn", "B", "", "DN of user for LDAP bind or use Env LDAP_BIND_DN")
	ldapCmd.PersistentFlags().StringVarP(&ldapBindPassword, "ldap.bindpassword", "p", "", "password for LDAP Bind User or use Env LDAP_BIND_PASSWORD")
	ldapCmd.PersistentFlags().BoolVar(&ldapTLS, "ldap.tls", false, "use secure ldap (ldaps)")
	ldapCmd.PersistentFlags().BoolVarP(&ldapInsecure, "ldap.insecure", "I", false, "do not verify TLS")
	ldapCmd.PersistentFlags().IntVarP(&ldapTimeout, "ldap.timeout", "t", ldapTimeout, "ldap timeout in sec")
	// ldapCmd.SetHelpFunc(hideFlags)
	ldapPassCmd.Flags().StringP("new-password", "n", "", "new_password to set or use Env LDAP_NEW_PASSWORD or be prompted")
	ldapPassCmd.Flags().BoolP("generate", "g", false, "generate a new password (alternative to be prompted)")
	ldapPassCmd.Flags().String("profile", "", "set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'")
	ldapPassCmd.Flags().String("profileset", "", "set profile to existing named profile set")
	ldapPassCmd.Flags().String("password_profiles", "", "filename for loading password profiled")
	ldapPassCmd.MarkFlagsMutuallyExclusive("new-password", "generate")
	ldapCmd.AddCommand(ldapPassCmd)

	ldapSSHCmd.Flags().StringP("sshpubkeyfile", "f", "id_rsa.pub", "filename with ssh public key to upload")
	ldapSSHCmd.SetHelpFunc(hideFlags)
	ldapCmd.AddCommand(ldapSSHCmd)

	ldapShowCmd.Flags().StringP("attributes", "A", "*", "comma separated list of attributes to show")
	ldapShowCmd.SetHelpFunc(hideFlags)
	ldapCmd.AddCommand(ldapShowCmd)

	ldapGroupCmd.PersistentFlags().StringVarP(&ldapGroupBase, "ldap.groupbase", "G", "", "Base DN for group search")
	ldapGroupCmd.SetHelpFunc(hideFlags)
	ldapCmd.AddCommand(ldapGroupCmd)

	ldapMemberCmd.PersistentFlags().StringVarP(&ldapGroupBase, "ldap.groupbase", "G", "", "Base DN for group membership search")
	ldapMemberCmd.Flags().StringP("group", "g", "", "group name or part to find their members")
	ldapMemberCmd.SetHelpFunc(hideFlags)
	ldapCmd.AddCommand(ldapMemberCmd)

	RootCmd.AddCommand(ldapCmd)

	if err := viper.BindPFlags(ldapCmd.PersistentFlags()); err != nil {
		log.Fatal(err)
	}
}

func initLdapConfig() {
	if ldapServer == "" {
		ldapServer = viper.GetString("ldap.host")
	}
	if ldapPort == 0 {
		ldapPort = viper.GetInt("ldap.port")
	}
	if ldapBaseDN == "" {
		ldapBaseDN = viper.GetString("ldap.base")
	}
	if ldapBindDN == "" {
		ldapBindDN = viper.GetString("ldap.binddn")
	}
	if ldapBindPassword == "" {
		ldapBindPassword = viper.GetString("ldap.bindpassword")
	}
	if ldapGroupBase == "" {
		ldapGroupBase = viper.GetString("ldap.groupbase")
	}

	if targetDN == "" {
		targetDN = ldapBindDN
	}

	if common.CmdFlagChanged(ldapCmd, "ldap.tls") {
		viper.Set("ldap.tls", ldapTLS)
	} else {
		ldapTLS = viper.GetBool("ldap.tls")
	}
	if common.CmdFlagChanged(ldapCmd, "ldap.insecure") {
		viper.Set("ldap.insecure", ldapInsecure)
	} else {
		ldapInsecure = viper.GetBool("ldap.insecure")
	}
	if common.CmdFlagChanged(ldapCmd, "ldap.timeout") {
		viper.Set("ldap.timeout", ldapTimeout)
	} else {
		ldapTimeout = viper.GetInt("ldap.timeout")
	}
}
func loadFromEnv() {
	if ldapBindDN == "" {
		ldapBindDN = os.Getenv("LDAP_BIND_DN")
		if ldapBindDN != "" {
			log.Debugf("use new LDAP_BIND_DN from env: %s", ldapBindDN)
		}
	}
	if ldapBindPassword == "" {
		ldapBindPassword = os.Getenv("LDAP_BIND_PASSWORD")
		if ldapBindPassword != "" {
			log.Debugf("use LDAP_BIND_PASSWORD from env")
		}
	}
}
func ldapLogin() (lc *ldaplib.LdapConfigType, err error) {
	initLdapConfig()
	loadFromEnv()
	if ldapBindDN == "" {
		err = fmt.Errorf("no LDAP Bind DN given, use --ldap.binddn or Env LDAP_BIND_DN")
		return
	}
	if ldapBaseDN == "" {
		p := strings.Split(ldapBindDN, ",")
		l := len(p)
		if l > 2 {
			ldapBaseDN = strings.Join(p[l-2:], ",")
			log.Debugf("use baseDN from bindDN: %s", ldapBaseDN)
		}
	}

	// query password if not given
	if ldapBindPassword == "" {
		pw := ""
		pw, err = promptPassword("Enter Bind password")
		if err != nil {
			err = fmt.Errorf("error reading password: %v", err)
			return
		}
		ldapBindPassword = pw
	}
	if ldapBindPassword == "" {
		err = fmt.Errorf("no LDAP Bind Password given, use --ldap.bindpass or Env LDAP_BIND_PASSWORD")
		return
	}

	log.Debugf("Try to connect to Ldap Server %s, Port %d, TLS %v, Insecure %v", ldapServer, ldapPort, ldapTLS, ldapInsecure)
	lc = ldaplib.NewConfig(ldapServer, ldapPort, ldapTLS, ldapInsecure, ldapBaseDN, ldapTimeout)
	err = lc.Connect(ldapBindDN, ldapBindPassword)
	if err != nil {
		err = fmt.Errorf("ldap bind to %s returned error %v", ldapBindDN, err)
		return
	}
	if lc.Conn != nil {
		log.Debugf("Ldap Connected")
	}
	return
}
