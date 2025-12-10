/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Specify the version of the application",
	Long:  `Prints the current version of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(pathfinder.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
