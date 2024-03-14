/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

const VERSION = "1.1.0"

const (
	sourceOrgFlagName   = "source-org"
	targetOrgFlagName   = "target-org"
	sourceTokenFlagName = "source-token"
	targetTokenFlagName = "target-token"
	maxRetriesFlagName  = "max-retries"
	workersFlagName     = "workers"
)

//go:embed banner.txt
var banner []byte

var enableDebug bool

func initLogger(cmd *cobra.Command, args []string) {
	lvl := new(slog.LevelVar)

	if enableDebug {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(slog.LevelInfo)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "gei-migration-helper",
	PersistentPreRun: initLogger,
	Short:            "Wrapper application to the GEI extension that orchestrates steps necessary to migrate reposistories and GHAS features",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.Version = VERSION
	rootCmd.Println("\n\n" + string(banner) + "\n\n")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().BoolVar(&enableDebug, "debug", os.Getenv("DEBUG") == "true", "Enable debug mode")

	rootCmd.PersistentFlags().String(sourceOrgFlagName, "", "The source organization.")
	rootCmd.MarkPersistentFlagRequired(sourceOrgFlagName)

	rootCmd.PersistentFlags().String(targetOrgFlagName, "", "The target organization.")
	rootCmd.MarkPersistentFlagRequired(targetOrgFlagName)

	rootCmd.PersistentFlags().String(sourceTokenFlagName, "", "The token of the source organization.")
	rootCmd.MarkPersistentFlagRequired(sourceTokenFlagName)

	rootCmd.PersistentFlags().String(targetTokenFlagName, "", "The token of the target organization.")
	rootCmd.MarkPersistentFlagRequired(targetTokenFlagName)

	rootCmd.PersistentFlags().Int(maxRetriesFlagName, 5, "[OPTIONAL] The maximum number of retries for a failed operation. Default: 5")
	rootCmd.PersistentFlags().Int(workersFlagName, 5, "[OPTIONAL] The number of workers to use for parallel operations. Default: 5")
}
