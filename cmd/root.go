package cmd

import (
	"log"
	"search/internal"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:     "fs",
	Short:   "Quick search for files",
	Version: "1.0.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		internal.ConnectDb()
	},
}

func Execute() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
