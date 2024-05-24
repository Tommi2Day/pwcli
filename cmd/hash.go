package cmd

import (
	//nolint: gosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/pwlib"
	"golang.org/x/crypto/bcrypt"
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "command to hashing Passwords ",
	Long: `prepare a password hash
currently supports md5 and scram(for postgresql),SSHA(vor LDAP) and bcrypt(for htpasswd) methods`,
	RunE:         doHash,
	SilenceUsage: true,
}

const mMD5 = "md5"
const mScram = "scram"
const mBcrypt = "bcrypt"
const mSSHA = "ssha"

func init() {
	hashCmd.Flags().StringP("username", "u", "", "username for scram and md5 hash")
	hashCmd.Flags().StringP("password", "p", "", "password to encrypt")
	hashCmd.Flags().StringP("prefix", "P", "", "prefix for hash string(default md5={MD5},ssha={SSHA})")
	hashCmd.Flags().StringP("test", "T", "", "test given hash to verify against hashed password (not for scram)")
	hashCmd.Flags().StringP("hash-method", "M", "", "method to use for hashing, one of md5, scram, bcrypt,ssha")
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
	case mSSHA:
		return hashSSHA(cmd, nil)
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
	prefix, _ := cmd.Flags().GetString("prefix")
	test, _ := cmd.Flags().GetString("test")
	if password == "" || username == "" {
		err = fmt.Errorf("username and password are required")
		return err
	}
	if prefix == "" && !cmd.Flags().Changed("prefix") {
		prefix = "{MD5}"
	}
	md5Value, err = doMD5(password + username)
	if err != nil {
		return fmt.Errorf("error while hashing password:%s", err)
	}
	result := prefix + md5Value
	log.Infof("MD5 hash for password '%s' is '%s'", password, result)
	if test == "" {
		cmd.Println(result)
		return nil
	}
	test = strings.TrimPrefix(test, prefix)
	test = strings.TrimPrefix(test, "md5")
	test = strings.TrimPrefix(test, "{MD5}")
	if test == md5Value {
		log.Infof("OK, test input matches md5 hash")
		cmd.Println("OK, test input matches md5 hash")
		return nil
	}
	log.Infof("ERROR, test input does not match md5 hash '%s'", md5Value)
	return fmt.Errorf("ERROR, test input does not match md5 hash '%s'", md5Value)
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
	test, _ := cmd.Flags().GetString("test")
	// generate only if test is empty
	if test == "" {
		bcryptValue, err = doBcrypt(password)
		if err != nil {
			return fmt.Errorf("error while hashing password:%s", err)
		}
		log.Infof("BCRYPT hash for password '%s' is '%s'", password, bcryptValue)
		cmd.Println(bcryptValue)
		return nil
	}
	// compare only if test is not empty
	err = bcrypt.CompareHashAndPassword([]byte(test), []byte(password))
	if err != nil {
		return fmt.Errorf("bcrypt compare failed:%s", err)
	}
	log.Infof("OK, test input matches bcrypt hash")
	cmd.Println("OK, test input matches bcrypt hash")
	return nil
}
func doBcrypt(password string) (hash string, err error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	return string(passwordBytes), nil
}

func hashSSHA(cmd *cobra.Command, _ []string) error {
	var err error
	var hash string
	password, _ := cmd.Flags().GetString("password")
	prefix, _ := cmd.Flags().GetString("prefix")
	test, _ := cmd.Flags().GetString("test")
	if password == "" {
		err = fmt.Errorf("password is required")
		return err
	}
	if prefix == "" && !cmd.Flags().Changed("prefix") {
		prefix = pwlib.SSHAPrefix
	}
	hash, err = doSSHA(password, prefix)
	if err != nil {
		return err
	}
	log.Infof("SSHA hash is '%s'", hash)
	if test == "" {
		cmd.Println(hash)
		return nil
	}
	test = strings.TrimPrefix(test, prefix)
	enc := pwlib.SSHAEncoder{}
	m := enc.Matches([]byte(test), []byte(password))
	if m {
		log.Infof("OK, test input matches ssha hash")
		cmd.Println("OK, test input matches ssha hash")
		return nil
	}
	log.Infof("ERROR, test input does not match ssha hash '%s'", hash)
	return fmt.Errorf("ERROR, test input does not match ssha hash '%s'", hash)
}
func doSSHA(sshaPlain string, prefix string) (result string, err error) {
	enc := pwlib.SSHAEncoder{}
	encoded, err := enc.Encode([]byte(sshaPlain), prefix)
	result = string(encoded)
	return
}
