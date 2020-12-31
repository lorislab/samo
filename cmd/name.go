package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	nameCmd = &cobra.Command{
		Use:   "name",
		Short: "Show the project name",
		Long:  `Tasks to show the maven project name`,
		Run: func(cmd *cobra.Command, args []string) {
			_, p := readProjectOptions()
			fmt.Printf("%s\n", p.Name())
		},
		TraverseChildren: true,
	}
)

func init() {
	projectCmd.AddCommand(nameCmd)
}
