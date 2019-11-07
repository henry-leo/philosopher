// Package cmd ProteinProphet top level command
package cmd

import (
	"os"

	"github.com/nesvilab/philosopher/lib/ext/proteinprophet"
	"github.com/nesvilab/philosopher/lib/met"
	"github.com/nesvilab/philosopher/lib/msg"
	"github.com/nesvilab/philosopher/lib/sys"
	"github.com/spf13/cobra"
)

// proprophCmd represents the proproph command
var proprophCmd = &cobra.Command{
	Use:   "proteinprophet",
	Short: "Protein identification validation",
	//Long:  "Statistical validation of protein identification based on peptide assignment to MS/MS spectra\nProteinProphet v5.0",
	Run: func(cmd *cobra.Command, args []string) {

		m.FunctionInitCheckUp()

		msg.Executing("ProteinProphet ", Version)

		m = proteinprophet.Run(m, args)

		m.Serialize()

		// clean tmp
		met.CleanTemp(m.Temp)

		msg.Done()
		return
	},
}

func init() {

	if len(os.Args) > 1 && os.Args[1] == "proteinprophet" {

		m.Restore(sys.Meta())

		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Iprophet, "iprophet", "", false, "input is from iProphet")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.ExcludeZ, "excludezeros", "", false, "exclude zero probability entries")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Noprotlen, "noprotlen", "", false, "do not report protein length")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Protmw, "protmw", "", false, "get protein molecular weights")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Icat, "icat", "", false, "highlight peptide cysteines")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Glyc, "glyc", "", false, "highlight peptide N-glycosylation motif")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Fpkm, "fpkm", "", false, "model protein FPKM values")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.NonSP, "nonsp", "", false, "do not use NSP model")
		proprophCmd.Flags().IntVarP(&m.ProteinProphet.Minindep, "minindep", "", 0, "minimum percentage of independent peptides required for a protein")
		proprophCmd.Flags().Float64VarP(&m.ProteinProphet.Minprob, "minprob", "", 0.05, "PeptideProphet probabilty threshold")
		proprophCmd.Flags().IntVarP(&m.ProteinProphet.Maxppmdiff, "maxppmdiff", "", 20, "maximum peptide mass difference in ppm")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Accuracy, "accuracy", "", false, "equivalent to --minprob 0")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Normprotlen, "normprotlen", "", false, "normalize NSP using protein length")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Nogroupwts, "nogroupwts", "", false, "check peptide's protein weight against the threshold (default: check peptide's protein group weight against threshold)")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Instances, "instances", "", false, "use expected number of ion instances to adjust the peptide probabilities prior to NSP adjustment")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Delude, "delude", "", false, "do NOT use peptide degeneracy information when assessing proteins")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Nooccam, "nooccam", "", false, "non-conservative maximum protein list")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Softoccam, "softoccam", "", false, "peptide weights are apportioned equally among proteins within each protein group (less conservative protein count estimate)")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Confem, "confem", "", false, "use the EM to compute probability given the confidence")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Logprobs, "logprobs", "", false, "use the log of the probabilities in the confidence calculations")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Allpeps, "allpeps", "", false, "consider all possible peptides in the database in the confidence model")
		proprophCmd.Flags().IntVarP(&m.ProteinProphet.Mufactor, "mufactor", "", 1, "fudge factor to scale MU calculation")
		proprophCmd.Flags().BoolVarP(&m.ProteinProphet.Unmapped, "unmapped", "", false, "report results for UNMAPPED proteins")
		proprophCmd.Flags().StringVarP(&m.ProteinProphet.Output, "output", "", "interact", "Output name")
		filterCmd.Flags().MarkHidden("accuracy")
		filterCmd.Flags().MarkHidden("allpeps")
		filterCmd.Flags().MarkHidden("confem")
		filterCmd.Flags().MarkHidden("delude")
		filterCmd.Flags().MarkHidden("excludezeros")
		filterCmd.Flags().MarkHidden("fpkm")
		filterCmd.Flags().MarkHidden("glyc")
		filterCmd.Flags().MarkHidden("icat")
		filterCmd.Flags().MarkHidden("instances")
		filterCmd.Flags().MarkHidden("logprobs")
		filterCmd.Flags().MarkHidden("minindep")
		filterCmd.Flags().MarkHidden("mufactor")
		filterCmd.Flags().MarkHidden("nogroupwts")
		filterCmd.Flags().MarkHidden("nooccam")
		filterCmd.Flags().MarkHidden("noprotlen")
		filterCmd.Flags().MarkHidden("normprotlen")
		filterCmd.Flags().MarkHidden("protmw")
		filterCmd.Flags().MarkHidden("softoccam")

	}

	RootCmd.AddCommand(proprophCmd)
}
