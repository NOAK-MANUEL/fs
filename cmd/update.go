package cmd

import (
	"search/internal"

	"github.com/spf13/cobra"
)

var update = &cobra.Command{
	Use:   "update",
	Short: "Update Synchronized files",
	Run: func(cmd *cobra.Command, args []string) {
		println("Updating...")
		internal.UpdateData()
	},
}

func init() {
	root.AddCommand(update)
}
