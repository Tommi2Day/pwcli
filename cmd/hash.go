package cmd

import (
	//nolint: gosec
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/matthewhartstonge/argon2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/pwlib"
	"golang.org/x/crypto/bcrypt"
)

const mMD5 = "md5"
const mScram = "scram"
const mBcrypt = "bcrypt"
const mSSHA = "ssha"
const mBasic = "basic"
const mArgon2 = "argon2"

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "command to hashing Passwords ",
	Long: `prepare a password hash
currently supports basic auth(for http), md5 and scram(for postgresql),SSHA(for LDAP), bcrypt(for htpasswd) and argon2(for vaultwarden)`,
}

var md5Cmd = &cobra.Command{
	Use:          mMD5,
	Short:        "command to hashing User/Password with MD5 method",
	RunE:         hashMD5,
	SilenceUsage: true,
}
var scramCmd = &cobra.Command{
	Use:          mScram,
	Short:        "command to hashing User/Password with SCRAM method",
	RunE:         hashScram,
	SilenceUsage: true,
}

var bcryptCmd = &cobra.Command{
	Use:          mBcrypt,
	Short:        "command to hashing Passwords with BCrypt method",
	RunE:         hashBcrypt,
	SilenceUsage: true,
}

var sshaCmd = &cobra.Command{
	Use:          mSSHA,
	Short:        "command to hashing Passwords with SSHA method",
	RunE:         hashSSHA,
	SilenceUsage: true,
}

var basicCmd = &cobra.Command{
	Use:          mBasic,
	Short:        "command to encoding User/Password with HTTP Basic method",
	RunE:         basicAuth,
	SilenceUsage: true,
}

var argon2Cmd = &cobra.Command{
	Use:          mArgon2,
	Short:        "command to hashing Passwords with Argon2 method",
	RunE:         hashArgon2,
	SilenceUsage: true,
}

func init() {
	// hide unused flags, do not on group command
	hashCmd.SetHelpFunc(hideFlags)
	RootCmd.AddCommand(hashCmd)

	md5Cmd.Flags().StringP("username", "u", "", "username ")
	md5Cmd.Flags().StringP("password", "p", "", "password to encrypt")
	md5Cmd.Flags().StringP("prefix", "P", "", "prefix for string(default md5={MD5}")
	md5Cmd.Flags().StringP("test", "T", "", "test given hash to verify against hashed password")
	_ = md5Cmd.MarkFlagRequired("password")
	_ = md5Cmd.MarkFlagRequired("username")
	hashCmd.AddCommand(md5Cmd)

	scramCmd.Flags().StringP("username", "u", "", "username")
	scramCmd.Flags().StringP("password", "p", "", "password to encrypt")
	scramCmd.Flags().StringP("prefix", "P", "", "prefix for hash string(default basic='Authorization: Basic ',md5={MD5},ssha={SSHA})")
	_ = scramCmd.MarkFlagRequired("password")
	_ = scramCmd.MarkFlagRequired("username")
	hashCmd.AddCommand(scramCmd)

	basicCmd.Flags().StringP("username", "u", "", "username")
	basicCmd.Flags().StringP("password", "p", "", "password to encode")
	basicCmd.Flags().StringP("prefix", "P", "", "prefix for hash string(default basic='Authorization: Basic ')")
	basicCmd.Flags().StringP("test", "T", "", "test given hash to verify against encoded password")
	_ = basicCmd.MarkFlagRequired("password")
	_ = basicCmd.MarkFlagRequired("username")
	hashCmd.AddCommand(basicCmd)

	bcryptCmd.Flags().StringP("password", "p", "", "password to encode")
	bcryptCmd.Flags().StringP("test", "T", "", "test given hash to verify against encoded password")
	_ = bcryptCmd.MarkFlagRequired("password")
	hashCmd.AddCommand(bcryptCmd)

	sshaCmd.Flags().StringP("password", "p", "", "password to encode")
	sshaCmd.Flags().StringP("test", "T", "", "test given hash to verify against encoded password")
	_ = md5Cmd.MarkFlagRequired("password")
	hashCmd.AddCommand(sshaCmd)

	argon2Cmd.Flags().StringP("password", "p", "", "password to encode")
	argon2Cmd.Flags().StringP("test", "T", "", "test given hash to verify against encoded password")
	_ = argon2Cmd.MarkFlagRequired("password")
	hashCmd.AddCommand(argon2Cmd)
}

func basicAuth(cmd *cobra.Command, _ []string) error {
	var err error
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	test, _ := cmd.Flags().GetString("test")
	prefix, _ := cmd.Flags().GetString("prefix")
	if password == "" || username == "" {
		err = fmt.Errorf("\"username and password are required")
		return err
	}
	if prefix == "" && !cmd.Flags().Changed("prefix") {
		prefix = "Authorization: Basic "
	}
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	result := auth
	if prefix != "" {
		result = prefix + auth
	}
	log.Infof("%s", result)

	if test == "" {
		cmd.Println(result)
		return nil
	}
	test = strings.TrimPrefix(test, prefix)
	if test == auth {
		log.Infof("OK, test input matches basic auth")
		cmd.Println("OK, test input matches basic")
		return nil
	}
	log.Infof("ERROR, test input '%s' does not match '%s'", test, auth)
	return fmt.Errorf("ERROR, test input does not match '%s'", auth)
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

func hashArgon2(cmd *cobra.Command, _ []string) error {
	var err error
	var argon2Value string
	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		err = fmt.Errorf("password is required")
		return err
	}
	test, _ := cmd.Flags().GetString("test")
	// generate only if test is empty
	if test == "" {
		argon2Value, err = doArgon2(password)
		if err != nil {
			return fmt.Errorf("error while hashing password:%s", err)
		}
		log.Infof("ARGON2 hash for password '%s' is '%s'", password, argon2Value)
		cmd.Println(argon2Value)
		return nil
	}
	// compare only if test is not empty
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(test))
	if err != nil {
		return fmt.Errorf("argon2 compare failed:%s", err)
	}

	if !ok {
		log.Infof("ERROR, test input does not match argon2 hash '%s'", argon2Value)
		return fmt.Errorf("ERROR, test input does not match ssha hash '%s'", argon2Value)
	}
	log.Infof("OK, test input matches argon2 hash")
	cmd.Println("OK, test input matches argon2 hash")
	return nil
}

func doArgon2(password string) (result string, err error) {
	argon := argon2.DefaultConfig()
	// Waaahht??! It includes magic salt generation for me ! Yasss...
	encoded, err := argon.HashEncoded([]byte(password))
	if err == nil {
		result = string(encoded)
	}
	return
}
