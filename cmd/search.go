package cmd

import (
	"search/internal"

	"github.com/spf13/cobra"
)

var withPath bool
var withExt bool
var search = &cobra.Command{
	Use:   "search",
	Short: "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		name := args[0]
		for _, path := range internal.Search(name, withPath) {
			println(path)
		}
	},
}

func init() {
	search.Flags().BoolVar(&withPath, "withPath", false, "search by path")
	search.Flags().BoolVar(&withExt, "withExt", false, "search by extension")
	root.AddCommand(search)
}
