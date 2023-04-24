/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const VERSION = "0.1.0"

const (
	sourceOrgFlagName  = "source-org"
	targetOrgFlagName  = "target-org"
	sourceTokenFlagName = "source-token"
	targetTokenFlagName = "target-token"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gei-migration-helper",
	Short: "Helper Application to prepare for GEI Migration",
	Long: `This CLI application helps to prepare for GEI Migration.
	It can be used to change the visibility of repositories, change GHAS settings for an organization and more.`,
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().String(sourceOrgFlagName, "", "The source organization.")
	// rootCmd.MarkFlagRequired(sourceOrgFlagName)

	// rootCmd.PersistentFlags().String(targetOrgFlagName, "", "The target organization.")
	// rootCmd.MarkFlagRequired(targetOrgFlagName)

	// rootCmd.PersistentFlags().String(sourceTokenFlagName, "", "The token of the source organization.")
	// rootCmd.MarkFlagRequired(sourceTokenFlagName)

	// rootCmd.PersistentFlags().String(targetTokenFlagName, "", "The token of the target organization.")
	// rootCmd.MarkFlagRequired(targetTokenFlagName)


}


