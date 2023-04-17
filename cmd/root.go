/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)


const (
	orgFlagName        = "org"
	repositoryFlagName = "repo"
	visibilityFlagName = "visibility"
	activateFlagName   = "activate"
	tokenFlagName	   = "token"
	deleteFlagName	   = "delete"
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gei-migration-helper.yaml)")

}


