package cmd

import "github.com/tommi2day/gomodules/common"

var kmsKeyID = common.GetStringEnv("KMS_KEYID", "")
var kmsEndpoint = common.GetStringEnv("KMS_ENDPOINT", "")
