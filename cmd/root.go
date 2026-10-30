package cmd

import (
	"fs/internal"
	"log"

	"github.com/spf13/cobra"
)

var withPath bool
var update bool

var root = &cobra.Command{
	Use:     "fs",
	Short:   "Quick search for files",
	Version: "1.0.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		internal.ConnectDb()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if update {
			println("Updating...")
			internal.UpdateData()
		} else {
			if len(args) < 1 {
				cmd.Usage()
				return
			}

			name := args[0]
			for _, path := range internal.Search(name, withPath) {
				println(path)
			}
		}

	},
}

func init() {
	root.Flags().BoolVar(&withPath, "withPath", false, "search by path")
	root.Flags().BoolVar(&update, "update", false, "Update Synchronized files")

}

func Execute() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
