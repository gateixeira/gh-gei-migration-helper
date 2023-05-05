/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"os"

	"github.com/spf13/cobra"
)

const VERSION = "0.1.2"

const (
	sourceOrgFlagName   = "source-org"
	targetOrgFlagName   = "target-org"
	sourceTokenFlagName = "source-token"
	targetTokenFlagName = "target-token"
)

//go:embed banner.txt
var banner []byte

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gei-migration-helper",
	Short: "Wrapper application to the GEI extension that orchestrates steps necessary to migrate reposistories and GHAS features",
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

	rootCmd.PersistentFlags().String(sourceOrgFlagName, "", "The source organization.")
	rootCmd.MarkFlagRequired(sourceOrgFlagName)

	rootCmd.PersistentFlags().String(targetOrgFlagName, "", "The target organization.")
	rootCmd.MarkFlagRequired(targetOrgFlagName)

	rootCmd.PersistentFlags().String(sourceTokenFlagName, "", "The token of the source organization.")
	rootCmd.MarkFlagRequired(sourceTokenFlagName)

	rootCmd.PersistentFlags().String(targetTokenFlagName, "", "The token of the target organization.")
	rootCmd.MarkFlagRequired(targetTokenFlagName)

}
