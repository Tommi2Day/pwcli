package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tommi2day/gomodules/common"
	"github.com/tommi2day/gomodules/pwlib"
)

var rdsRegion = common.GetStringEnv("RDS_REGION", "")
var rdsPort = pwlib.RDSDefaultPort

var rdsCmd = &cobra.Command{
	Use:   typeRDS,
	Short: "handle AWS RDS IAM authentication",
	Long:  `Generate AWS RDS IAM authentication tokens to be used as database password`,
}

var rdsTokenCmd = &cobra.Command{
	Use:   cmdToken,
	Short: "generate an RDS IAM auth token",
	Long: `
generate an IAM authentication token for the given RDS endpoint (host[:port]) and database user.
The token is valid for 15 minutes and is used as password for the database login`,
	RunE:         rdsToken,
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(rdsCmd)
	rdsCmd.PersistentFlags().StringVar(&rdsRegion, "rds_region", rdsRegion, "AWS region of the RDS instance, default region of the AWS config (RDS_REGION)")
	rdsCmd.PersistentFlags().StringVar(&rdsPort, "rds_port", rdsPort, "RDS port used if the endpoint contains no port")

	hideGlobalFlags(rdsTokenCmd)
	rdsTokenCmd.Flags().StringP("endpoint", "H", "", "RDS endpoint host[:port]")
	rdsTokenCmd.Flags().StringP("user", "u", "", "database user")
	rdsCmd.AddCommand(rdsTokenCmd)
}

// setRDSParams propagates the --rds_region and --rds_port flags to pwlib.
// The region is set as env var, because pwlib prefers RDS_REGION over pwlib.RDSRegion.
func setRDSParams() {
	if rdsRegion != "" {
		log.Debugf("use RDS region %s", rdsRegion)
		_ = os.Setenv("RDS_REGION", rdsRegion)
	}
	if rdsPort != "" {
		pwlib.RDSDefaultPort = rdsPort
	}
}

func rdsToken(cmd *cobra.Command, _ []string) error {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	user, _ := cmd.Flags().GetString("user")
	if endpoint == "" || user == "" {
		return fmt.Errorf("rds token needs parameter endpoint and user set")
	}
	log.Debugf("RDS token for '%s'@'%s' entered", user, endpoint)
	setRDSParams()
	token, err := pwlib.GetRDSAuthToken(endpoint, "", user)
	if err != nil {
		return err
	}
	fmt.Println(token)
	log.Info("RDS auth token successfully generated")
	return nil
}
