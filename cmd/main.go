package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/jessevdk/go-flags"
	"gitlab.com/weeve/edge-server/edge-pipeline-service/internal"
	// "gitlab.com/weeve/edge-server/edge-pipeline-service/internal/constants"
)

type Options struct {
	Port    int    `long:"port" short:"p" description:"Port number" required:"true"`
	Verbose []bool `long:"verbose" short:"v" description:"Show verbose debug information"`

	// TODO: We only need this for AWS ECR integration...
	// RoleArn string `long:"role" short:"r" description:"Role Arn" required:"false"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
	log.Info("Started logging")
}

// @title Weeve Manager API
// @version 1.0
// @description This is a weeve management api.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Token
func main() {
	if _, err := parser.Parse(); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// Set logging level from -v flag
	if len(options.Verbose) >= 1 {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Info("Logging level set to ", log.GetLevel())

	// TODO: We only need this for AWS ECR integration...
	// if options.RoleArn != "" {
	// 	constants.RoleArn = options.RoleArn
	// }

	log.Info("Starting server on port ", options.Port)

	internal.HandleRequests(options.Port)
}
