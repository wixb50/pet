package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"pet/config"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Select the file edit",
	Long:  `Select the file edit directly`,
	RunE:  use,
}

func use(cmd *cobra.Command, args []string) (err error) {
	var useFile string

	if len(args) > 0 {
		useFile = args[0]
	} else {
		flag := config.Flag
		var options []string
		if flag.Query != "" {
			options = append(options, fmt.Sprintf("--query %s", flag.Query))
		}

		allFiles, err := filterSnippetFile(options)
		if err != nil {
			return err
		}
		useFile = allFiles[0]
	}

	if err := config.Conf.Switch(useFile); err != nil {
		return err
	}

	fmt.Printf("Use this file: %s\n", useFile)
	return nil
}

func init() {
	RootCmd.AddCommand(useCmd)
	useCmd.Flags().StringVarP(&config.Flag.Query, "query", "q", "",
		`Initial value for query`)
}
