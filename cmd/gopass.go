package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/pwlib"
	"gopkg.in/yaml.v3"
)

var (
	gopassStoreDir    string
	gopassCrypto      string
	gopassKeyFile     string
	gopassIdentityDir string
)

var gopassCmd = &cobra.Command{
	Use:   "gopass",
	Short: "Manage gopass password store",
	Long:  `Read, write and manage secrets in a gopass-compatible password store`,
}

var gopassListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List secrets in store",
	RunE:         gopassList,
	SilenceUsage: true,
}

var gopassReadCmd = &cobra.Command{
	Use:          "read <secret>",
	Short:        "Decrypt and print a secret",
	Args:         cobra.ExactArgs(1),
	RunE:         gopassRead,
	SilenceUsage: true,
}

var gopassWriteCmd = &cobra.Command{
	Use:          "write <secret>",
	Short:        "Encrypt and store a secret",
	Args:         cobra.ExactArgs(1),
	RunE:         gopassWrite,
	SilenceUsage: true,
}

var gopassStoresCmd = &cobra.Command{
	Use:          "stores",
	Short:        "List configured gopass stores from gopass config",
	RunE:         gopassStores,
	SilenceUsage: true,
}

var gopassRecipientsCmd = &cobra.Command{
	Use:   "recipients",
	Short: "Manage store recipients",
}

var gopassRecipientsListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List recipients in store (.age-recipients or .gpg-id)",
	RunE:         gopassRecipientsList,
	SilenceUsage: true,
}

var gopassRecipientsAddCmd = &cobra.Command{
	Use:          "add <pubkey>",
	Short:        "Append a public key to the recipients file",
	Args:         cobra.ExactArgs(1),
	RunE:         gopassRecipientsAdd,
	SilenceUsage: true,
}

var gopassIdentityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Manage age and GPG identity files",
}

var gopassIdentityListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List age identity files in identity directory",
	RunE:         gopassIdentityList,
	SilenceUsage: true,
}

var gopassIdentityAddCmd = &cobra.Command{
	Use:          "add <alias> <keyfile>",
	Short:        "Copy an age private key into the identity directory",
	Args:         cobra.ExactArgs(2),
	RunE:         gopassIdentityAdd,
	SilenceUsage: true,
}

var gopassIdentityCreateCmd = &cobra.Command{
	Use:          "create <alias>",
	Short:        "Generate a new age or GPG key pair and store it in the identity directory",
	Args:         cobra.ExactArgs(1),
	RunE:         gopassIdentityCreate,
	SilenceUsage: true,
}

var gopassPullCmd = &cobra.Command{
	Use:          "pull",
	Short:        "Git pull the store directory",
	RunE:         gopassPull,
	SilenceUsage: true,
}

var gopassPushCmd = &cobra.Command{
	Use:          "push",
	Short:        "Git push the store directory",
	RunE:         gopassPush,
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(gopassCmd)

	gopassCmd.PersistentFlags().StringVarP(&gopassStoreDir, "store-dir", "S", "", "Path to gopass store directory (auto-detected if empty)")
	gopassCmd.PersistentFlags().StringVarP(&gopassCrypto, "crypto", "C", "", "Encryption type: age or gpg (auto-detected if empty)")
	gopassCmd.PersistentFlags().StringVarP(&gopassKeyFile, "key-file", "k", "", "Age identity file (read) or recipients file (write)")
	gopassCmd.PersistentFlags().StringVar(&gopassIdentityDir, "identity-dir", "", "Age identity directory for auto-detection (default: ~/.config/gopass/identities/)")

	hideGlobalFlags(gopassListCmd)
	hideGlobalFlags(gopassReadCmd)
	hideGlobalFlags(gopassWriteCmd)
	hideGlobalFlags(gopassStoresCmd)
	hideGlobalFlags(gopassRecipientsListCmd)
	hideGlobalFlags(gopassRecipientsAddCmd)
	hideGlobalFlags(gopassIdentityListCmd)
	hideGlobalFlags(gopassIdentityAddCmd)
	hideGlobalFlags(gopassIdentityCreateCmd)
	hideGlobalFlags(gopassPullCmd)
	hideGlobalFlags(gopassPushCmd)

	gopassReadCmd.Flags().Bool("raw", false, "Output full raw secret content instead of first line only")
	gopassWriteCmd.Flags().String("content", "", "Secret content to store (reads from stdin if not set)")

	gopassIdentityCreateCmd.Flags().String("name", "", "GPG identity name")
	gopassIdentityCreateCmd.Flags().String("email", "", "GPG identity email")
	gopassIdentityCreateCmd.Flags().String("comment", "", "GPG identity comment")
	gopassIdentityCreateCmd.Flags().String("passphrase", "", "GPG private key passphrase")
	gopassIdentityCreateCmd.Flags().Bool("add-recipient", false, "Append the new public key to the store recipients file")

	gopassPullCmd.Flags().String("remote", "origin", "Git remote name")
	gopassPushCmd.Flags().String("remote", "origin", "Git remote name")

	gopassCmd.AddCommand(gopassListCmd)
	gopassCmd.AddCommand(gopassReadCmd)
	gopassCmd.AddCommand(gopassWriteCmd)
	gopassCmd.AddCommand(gopassStoresCmd)

	gopassRecipientsCmd.AddCommand(gopassRecipientsListCmd)
	gopassRecipientsCmd.AddCommand(gopassRecipientsAddCmd)
	gopassCmd.AddCommand(gopassRecipientsCmd)

	gopassIdentityCmd.AddCommand(gopassIdentityListCmd)
	gopassIdentityCmd.AddCommand(gopassIdentityAddCmd)
	gopassIdentityCmd.AddCommand(gopassIdentityCreateCmd)
	gopassCmd.AddCommand(gopassIdentityCmd)

	gopassCmd.AddCommand(gopassPullCmd)
	gopassCmd.AddCommand(gopassPushCmd)
}

func resolveGopassCrypto(storeDir, flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	return pwlib.GopassDetectCrypto(storeDir)
}

// gopassAgeConfigSection captures the optional [age] section of the gopass
// config YAML that pwlib's GopassConfig does not expose.
type gopassAgeConfigSection struct {
	Age struct {
		IdentityDir string `yaml:"identity_dir"`
	} `yaml:"age"`
}

// gopassIdentityDirFromConfigFile reads the gopass config file at configPath
// and returns the value of age.identity_dir if present, or "".
func gopassIdentityDirFromConfigFile(configPath string) string {
	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return ""
	}
	var cfg gopassAgeConfigSection
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.Age.IdentityDir
}

// resolveGopassIdentityDir resolves the age identity directory in this order:
//  1. --identity-dir flag (flagValue)
//  2. GOPASS_IDENTITY_DIR environment variable
//  3. age.identity_dir field in the gopass config file
//  4. identities/ sibling of the gopass config file
//     (inherits GOPASS_CONFIG and XDG_CONFIG_HOME automatically via pwlib)
//  5. ~/.config/gopass/identities/ absolute fallback
func resolveGopassIdentityDir(flagValue string) (string, error) {
	if flagValue != "" {
		log.Debugf("gopass identity dir from flag: %s", flagValue)
		return flagValue, nil
	}
	if dir := os.Getenv("GOPASS_IDENTITY_DIR"); dir != "" {
		log.Debugf("gopass identity dir from GOPASS_IDENTITY_DIR: %s", dir)
		return dir, nil
	}
	if configPath, err := pwlib.GopassConfigPath(); err == nil {
		if dir := gopassIdentityDirFromConfigFile(configPath); dir != "" {
			log.Debugf("gopass identity dir from config file field: %s", dir)
			return dir, nil
		}
		dir := filepath.Join(filepath.Dir(configPath), "identities")
		log.Debugf("gopass identity dir derived from config path: %s", dir)
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "gopass", "identities"), nil
}

// gopassFindIdentity scans the identity directory for a *.key file that can
// decrypt secretName in storeDir, using pwlib.AgeDetectIdentity.
func gopassFindIdentity(storeDir, secretName string) (string, error) {
	identityDir, err := resolveGopassIdentityDir(gopassIdentityDir)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(identityDir)
	if err != nil {
		return "", fmt.Errorf("identity dir %s not readable: %w", identityDir, err)
	}
	var keyFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".key") {
			keyFiles = append(keyFiles, filepath.Join(identityDir, e.Name()))
		}
	}
	if len(keyFiles) == 0 {
		return "", fmt.Errorf("no identity files found in %s; use --key-file or --identity-dir", identityDir)
	}
	encFile := filepath.Join(storeDir, filepath.FromSlash(secretName)+".age")
	matched, err := pwlib.AgeDetectIdentity(encFile, keyFiles)
	if err != nil {
		return "", fmt.Errorf("auto-detect identity for %s: %w", secretName, err)
	}
	log.Debugf("auto-detected identity %s for secret %s", matched, secretName)
	return matched, nil
}

func gopassResolveStore() (storeDir, cryptoType string, err error) {
	storeDir, err = pwlib.GopassStoreDir(gopassStoreDir)
	if err != nil {
		return
	}
	cryptoType, err = resolveGopassCrypto(storeDir, gopassCrypto)
	return
}

func gopassList(cmd *cobra.Command, _ []string) error {
	storeDir, cryptoType, err := gopassResolveStore()
	if err != nil {
		return err
	}
	log.Debugf("gopass list storeDir=%s crypto=%s", storeDir, cryptoType)
	secrets, err := pwlib.GopassList(storeDir, cryptoType)
	if err != nil {
		return err
	}
	for _, s := range secrets {
		cmd.Println(s)
	}
	return nil
}

func gopassRead(cmd *cobra.Command, args []string) error {
	storeDir, cryptoType, err := gopassResolveStore()
	if err != nil {
		return err
	}
	keyFile := gopassKeyFile
	if keyFile == "" && cryptoType == pwlib.GopassCryptoAge {
		keyFile, err = gopassFindIdentity(storeDir, args[0])
		if err != nil {
			return err
		}
	}
	raw, _ := cmd.Flags().GetBool("raw")
	log.Debugf("gopass read secret=%s storeDir=%s crypto=%s raw=%v keyFile=%s", args[0], storeDir, cryptoType, raw, keyFile)
	var content string
	if raw {
		content, err = pwlib.GopassReadRaw(storeDir, args[0], keyFile, "", cryptoType)
	} else {
		content, err = pwlib.GopassRead(storeDir, args[0], keyFile, "", cryptoType)
	}
	if err != nil {
		return err
	}
	cmd.Println(content)
	return nil
}

func gopassWrite(cmd *cobra.Command, args []string) error {
	storeDir, cryptoType, err := gopassResolveStore()
	if err != nil {
		return err
	}
	content, _ := cmd.Flags().GetString("content")
	if content == "" {
		data, readErr := io.ReadAll(cmd.InOrStdin())
		if readErr != nil {
			return fmt.Errorf("reading stdin: %w", readErr)
		}
		content = string(data)
	}
	log.Debugf("gopass write secret=%s storeDir=%s crypto=%s", args[0], storeDir, cryptoType)
	if err = pwlib.GopassWrite(storeDir, args[0], content, gopassKeyFile, cryptoType); err != nil {
		return err
	}
	cmd.Printf("secret %s written\n", args[0])
	return nil
}

func gopassStores(cmd *cobra.Command, _ []string) error {
	cfg, err := pwlib.GopassReadConfig("")
	if err != nil {
		return err
	}
	cmd.Printf("root: %s [%s]\n", cfg.Root.Path, cfg.Root.Crypto)
	for name, mount := range cfg.Mounts {
		cmd.Printf("%s: %s [%s]\n", name, mount.Path, mount.Crypto)
	}
	return nil
}

func gopassRecipientsMarker(storeDir, cryptoType string) string {
	if cryptoType == pwlib.GopassCryptoAge {
		return filepath.Join(storeDir, ".age-recipients")
	}
	return filepath.Join(storeDir, ".gpg-id")
}

func gopassRecipientsList(cmd *cobra.Command, _ []string) error {
	storeDir, cryptoType, err := gopassResolveStore()
	if err != nil {
		return err
	}
	markerFile := gopassRecipientsMarker(storeDir, cryptoType)
	data, err := os.ReadFile(markerFile) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read recipients file %s: %w", markerFile, err)
	}
	cmd.Print(string(data))
	return nil
}

func gopassRecipientsAdd(cmd *cobra.Command, args []string) error {
	storeDir, cryptoType, err := gopassResolveStore()
	if err != nil {
		return err
	}
	markerFile := gopassRecipientsMarker(storeDir, cryptoType)
	f, err := os.OpenFile(markerFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open recipients file %s: %w", markerFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	if _, err = fmt.Fprintln(f, args[0]); err != nil {
		return fmt.Errorf("write to recipients file: %w", err)
	}
	cmd.Printf("recipient added to %s\n", markerFile)
	return nil
}

func gopassIdentityList(cmd *cobra.Command, _ []string) error {
	dir, err := resolveGopassIdentityDir(gopassIdentityDir)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("identity dir %s does not exist", dir)
			return nil
		}
		return fmt.Errorf("read identity dir %s: %w", dir, err)
	}
	for _, e := range entries {
		cmd.Println(e.Name())
	}
	return nil
}

func gopassIdentityAdd(cmd *cobra.Command, args []string) error {
	alias := args[0]
	srcKeyFile := args[1]
	dir, err := resolveGopassIdentityDir(gopassIdentityDir)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create identity dir %s: %w", dir, err)
	}
	data, err := os.ReadFile(srcKeyFile) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read key file %s: %w", srcKeyFile, err)
	}
	dest := filepath.Join(dir, alias+".key")
	if err = os.WriteFile(dest, data, 0600); err != nil { //nolint:gosec
		return fmt.Errorf("write identity file %s: %w", dest, err)
	}
	cmd.Printf("identity %s added as %s\n", alias, dest)
	return nil
}

func gopassIdentityCreate(cmd *cobra.Command, args []string) error {
	alias := args[0]

	// Resolve crypto: flag → auto-detect from store → default age
	cryptoType := gopassCrypto
	if cryptoType == "" {
		storeDir, sErr := pwlib.GopassStoreDir(gopassStoreDir)
		if sErr == nil {
			if detected, dErr := pwlib.GopassDetectCrypto(storeDir); dErr == nil {
				cryptoType = detected
			}
		}
	}
	if cryptoType == "" {
		cryptoType = pwlib.GopassCryptoAge
	}

	identityDir, err := resolveGopassIdentityDir(gopassIdentityDir)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(identityDir, 0700); err != nil {
		return fmt.Errorf("create identity dir %s: %w", identityDir, err)
	}

	privDest := filepath.Join(identityDir, alias+".key")
	pubDest := filepath.Join(identityDir, alias+".pub")

	var pubKeyStr string
	switch cryptoType {
	case pwlib.GopassCryptoAge:
		pubKeyStr, err = createAgeIdentityFiles(privDest, pubDest)
	default: // gpg
		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		comment, _ := cmd.Flags().GetString("comment")
		passphrase, _ := cmd.Flags().GetString("passphrase")
		pubKeyStr, err = createGPGIdentityFiles(name, comment, email, passphrase, privDest, pubDest)
	}
	if err != nil {
		return err
	}

	log.Debugf("identity %s created: priv=%s pub=%s crypto=%s", alias, privDest, pubDest, cryptoType)
	cmd.Printf("identity %s created [%s]\n", alias, cryptoType)
	cmd.Printf("  private key: %s\n", privDest)
	cmd.Printf("  public key:  %s\n", pubDest)
	cmd.Printf("  public key value: %s\n", pubKeyStr)

	addRecipient, _ := cmd.Flags().GetBool("add-recipient")
	if !addRecipient {
		return nil
	}
	storeDir, err := pwlib.GopassStoreDir(gopassStoreDir)
	if err != nil {
		return err
	}
	marker := gopassRecipientsMarker(storeDir, cryptoType)
	f, err := os.OpenFile(marker, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open recipients file %s: %w", marker, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	if _, err = fmt.Fprintln(f, pubKeyStr); err != nil {
		return fmt.Errorf("write to recipients file: %w", err)
	}
	cmd.Printf("  added to recipients: %s\n", marker)
	return nil
}

func createAgeIdentityFiles(privDest, pubDest string) (pubKeyStr string, err error) {
	identity, recipient, err := pwlib.CreateAgeIdentity()
	if err != nil {
		return "", fmt.Errorf("generate age identity: %w", err)
	}
	if err = pwlib.ExportAgeKeyPair(identity, pubDest, privDest); err != nil {
		return "", fmt.Errorf("export age key pair: %w", err)
	}
	return recipient, nil
}

func createGPGIdentityFiles(name, comment, email, passphrase, privDest, pubDest string) (pubKeyStr string, err error) {
	entity, keyID, err := pwlib.CreateGPGEntity(name, comment, email, passphrase)
	if err != nil {
		return "", fmt.Errorf("generate GPG identity: %w", err)
	}
	if err = pwlib.ExportGPGKeyPair(entity, pubDest, privDest); err != nil {
		return "", fmt.Errorf("export GPG key pair: %w", err)
	}
	return keyID, nil
}

func gopassPull(cmd *cobra.Command, _ []string) error {
	storeDir, err := pwlib.GopassStoreDir(gopassStoreDir)
	if err != nil {
		return err
	}
	remote, _ := cmd.Flags().GetString("remote")
	gitArgs := []string{"-C", storeDir, "pull"}
	if remote != "" {
		gitArgs = append(gitArgs, remote)
	}
	log.Debugf("git %v", gitArgs)
	out, err := exec.Command("git", gitArgs...).CombinedOutput() //nolint:gosec
	cmd.Print(string(out))
	if err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	return nil
}

func gopassPush(cmd *cobra.Command, _ []string) error {
	storeDir, err := pwlib.GopassStoreDir(gopassStoreDir)
	if err != nil {
		return err
	}
	remote, _ := cmd.Flags().GetString("remote")
	gitArgs := []string{"-C", storeDir, "push"}
	if remote != "" {
		gitArgs = append(gitArgs, remote)
	}
	log.Debugf("git %v", gitArgs)
	out, err := exec.Command("git", gitArgs...).CombinedOutput() //nolint:gosec
	cmd.Print(string(out))
	if err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}
