package cmd

import (
	//nolint: gosec
	"crypto/md5"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/pwlib"
	"golang.org/x/crypto/bcrypt"
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "command to hashing Passwords for use in Postgresql",
	Long: `prepare a password hash
currently supports md5 and scram(for postgresql) and bcrypt(for htpasswd) methods`,
	RunE:         doHash,
	SilenceUsage: true,
}

const mMD5 = "md5"
const mScram = "scram"
const mBcrypt = "bcrypt"

func init() {
	hashCmd.Flags().String("username", "", "username for scram encrypt")
	hashCmd.Flags().String("password", "", "password to encrypt")
	hashCmd.Flags().String("hash-method", "", "method to use for hashing, one of md5, scram, bcrypt")
	_ = hashCmd.MarkFlagRequired("password")
	_ = hashCmd.MarkFlagRequired("hash-method")
	hashCmd.SetHelpFunc(hideFlags)
	RootCmd.AddCommand(hashCmd)
	// hide unused flags, do not on group command
}

func doHash(cmd *cobra.Command, _ []string) error {
	m, _ := cmd.Flags().GetString("hash-method")
	switch m {
	case mMD5:
		return hashMD5(cmd, nil)
	case mScram:
		return hashScram(cmd, nil)
	case mBcrypt:
		return hashBcrypt(cmd, nil)
	default:
		return fmt.Errorf("method %s not supported", m)
	}
}

func hashScram(cmd *cobra.Command, _ []string) error {
	var err error
	var scramValue string
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	if password == "" || username == "" {
		err = fmt.Errorf("username and password are required")
		return err
	}
	scramValue, err = pwlib.ScramPassword(username, password)
	if err != nil {
		return err
	}
	log.Infof("SCRAM-SHA-256 hash is '%s'", scramValue)
	cmd.Println(scramValue)
	return nil
}

func hashMD5(cmd *cobra.Command, _ []string) error {
	var err error
	var md5Value string
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	if password == "" || username == "" {
		err = fmt.Errorf("username and password are required")
		return err
	}
	md5Value, err = doMD5(password + username)
	if err != nil {
		return fmt.Errorf("error while hashing password:%s", err)
	}
	md5Value = "md5" + md5Value
	log.Infof("MD5 hash for password '%s' is '%s'", password, md5Value)
	cmd.Println(md5Value)
	return nil
}

func doMD5(text string) (result string, err error) {
	//nolint: gosec
	h := md5.New()
	_, err = h.Write([]byte(text))
	if err != nil {
		return "", err
	}
	result = hex.EncodeToString(h.Sum(nil))
	return
}

func hashBcrypt(cmd *cobra.Command, _ []string) error {
	var err error
	var bcryptValue string
	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		err = fmt.Errorf("password is required")
		return err
	}
	bcryptValue, err = doBcrypt(password)
	if err != nil {
		return fmt.Errorf("error while hashing password:%s", err)
	}
	log.Infof("BCRYPT hash for password '%s' is '%s'", password, bcryptValue)
	cmd.Println(bcryptValue)
	return nil
}
func doBcrypt(password string) (hash string, err error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	return string(passwordBytes), nil
}
