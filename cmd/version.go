package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
)

var (
	shortened = false
	output    = "json"
	bv        BuildVersion
)

type BuildVersion struct {
	Version string
	Commit  string
	Date    string
}

func createVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
			resp := goVersion.FuncWithOutput(shortened, bv.Version, bv.Commit, bv.Date, output)
			fmt.Print(resp)
		},
	}

	cmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number.")
	cmd.Flags().StringVarP(&output, "output", "o", "json", "Output format. One of 'yaml' or 'json'.")

	return cmd
}
