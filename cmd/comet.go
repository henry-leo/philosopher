package cmd

import (
	"os"

	"github.com/prvst/philosopher/lib/err"
	"github.com/prvst/philosopher/lib/ext/comet"
	"github.com/prvst/philosopher/lib/met"
	"github.com/prvst/philosopher/lib/sys"
	"github.com/spf13/cobra"
)

// cometCmd represents the comet command
var cometCmd = &cobra.Command{
	Use:   "comet",
	Short: "Peptide spectrum matching with Comet",
	Run: func(cmd *cobra.Command, args []string) {

		m.FunctionInitCheckUp()

		err.Executing("Comet ", Version)

		m = comet.Run(m, args)

		m.Serialize()

		// clean tmp
		met.CleanTemp(m.Temp)

		err.Done()
		return
	},
}

func init() {

	if len(os.Args) > 1 && os.Args[1] == "comet" {

		m.Restore(sys.Meta())

		cometCmd.Flags().BoolVarP(&m.Comet.Print, "print", "", false, "print a comet.params file")
		cometCmd.Flags().BoolVarP(&m.Comet.NoIndex, "noindex", "", false, "skip raw file indexing")
		cometCmd.Flags().StringVarP(&m.Comet.Param, "param", "", "comet.params.txt", "comet parameter file")
	}

	RootCmd.AddCommand(cometCmd)
}
