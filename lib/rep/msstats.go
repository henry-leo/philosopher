package rep

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nesvilab/philosopher/lib/bio"
	"github.com/nesvilab/philosopher/lib/msg"
	"github.com/nesvilab/philosopher/lib/sys"
)

// MetaMSstatsReport report all psms from study that passed the FDR filter
func (evi Evidence) MetaMSstatsReport(labels map[string]string, brand string, channels int, hasDecoys bool) {

	var header string
	output := fmt.Sprintf("%s%smsstats.csv", sys.MetaDir(), string(filepath.Separator))

	// create result file
	file, e := os.Create(output)
	if e != nil {
		msg.WriteFile(errors.New("Cannot create MSstats report"), "error")
	}
	defer file.Close()

	// building the printing set tat may or not contain decoys
	var printSet PSMEvidenceList
	for _, i := range evi.PSM {
		if hasDecoys == false {
			if i.IsDecoy == false {
				printSet = append(printSet, i)
			}
		} else {
			printSet = append(printSet, i)
		}
	}

	header = "Spectrum.Name\tSpectrum.File\tPeptide.Sequence\tModified.Peptide.Sequence\tCharge\tCalculated.MZ\tPeptideProphet.Probability\tIntensity\tIs.Unique\tGene\tProtein.Accessions\tModifications"

	if brand == "tmt" {
		switch channels {
		case 10:
			header += "\tPurity\tChannel 126\tChannel 127N\tChannel 127C\tChannel 128N\tChannel 128C\tChannel 129N\tChannel 129C\tChannel 130N\tChannel 130C\tChannel 131N"
		case 11:
			header += "\tPurity\tChannel 126\tChannel 127N\tChannel 127C\tChannel 128N\tChannel 128C\tChannel 129N\tChannel 129C\tChannel 130N\tChannel 130C\tChannel 131N\tChannel 131C"
		case 16:
			header += "\tPurity\tChannel 126\tChannel 127N\tChannel 127C\tChannel 128N\tChannel 128C\tChannel 129N\tChannel 129C\tChannel 130N\tChannel 130C\tChannel 131N\tChannel 131C\tChannel 132N\tChannel 132C\tChannel 133N\tChannel 133C\tChannel 134N"
		default:
			header += ""
		}
	}

	header += "\n"

	if len(labels) > 0 {
		for k, v := range labels {
			k = fmt.Sprintf("Channel %s", k)
			header = strings.Replace(header, k, v, -1)
		}
	}

	_, e = io.WriteString(file, header)
	if e != nil {
		msg.WriteToFile(errors.New("Cannot print PSM to file"), "fatal")
	}

	for _, i := range printSet {

		var fileName string
		parts := strings.Split(i.Spectrum, ".")
		fileName = fmt.Sprintf("%s.raw", parts[0])

		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%.4f\t%.4f\t%.4f\t%t\t%s\t%s\t%s",
			i.Spectrum,
			fileName,
			i.Peptide,
			i.ModifiedPeptide,
			i.AssumedCharge,
			((i.CalcNeutralPepMass + (float64(i.AssumedCharge) * bio.Proton)) / float64(i.AssumedCharge)),
			i.Probability,
			i.Intensity,
			i.IsUnique,
			i.GeneName,
			i.Protein,
			"",
		)

		if brand == "tmt" {
			switch channels {
			case 10:
				line = fmt.Sprintf("%s\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f",
					line,
					i.Purity,
					i.Labels.Channel1.Intensity,
					i.Labels.Channel2.Intensity,
					i.Labels.Channel3.Intensity,
					i.Labels.Channel4.Intensity,
					i.Labels.Channel5.Intensity,
					i.Labels.Channel6.Intensity,
					i.Labels.Channel7.Intensity,
					i.Labels.Channel8.Intensity,
					i.Labels.Channel9.Intensity,
					i.Labels.Channel10.Intensity,
				)
			case 11:
				line = fmt.Sprintf("%s\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f",
					line,
					i.Purity,
					i.Labels.Channel1.Intensity,
					i.Labels.Channel2.Intensity,
					i.Labels.Channel3.Intensity,
					i.Labels.Channel4.Intensity,
					i.Labels.Channel5.Intensity,
					i.Labels.Channel6.Intensity,
					i.Labels.Channel7.Intensity,
					i.Labels.Channel8.Intensity,
					i.Labels.Channel9.Intensity,
					i.Labels.Channel10.Intensity,
					i.Labels.Channel11.Intensity,
				)
			case 16:
				line = fmt.Sprintf("%s\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f",
					line,
					i.Purity,
					i.Labels.Channel1.Intensity,
					i.Labels.Channel2.Intensity,
					i.Labels.Channel3.Intensity,
					i.Labels.Channel4.Intensity,
					i.Labels.Channel5.Intensity,
					i.Labels.Channel6.Intensity,
					i.Labels.Channel7.Intensity,
					i.Labels.Channel8.Intensity,
					i.Labels.Channel9.Intensity,
					i.Labels.Channel10.Intensity,
					i.Labels.Channel11.Intensity,
					//i.Labels.Channel12.Intensity,
					//i.Labels.Channel13.Intensity,
					//i.Labels.Channel14.Intensity,
					//i.Labels.Channel15.Intensity,
					//i.Labels.Channel16.Intensity,
				)
			default:
				header += ""
			}
		}

		line += "\n"

		_, e = io.WriteString(file, line)
		if e != nil {
			msg.WriteToFile(errors.New("Cannot write to MSstats report"), "fatal")
		}
	}

	// copy to work directory
	sys.CopyFile(output, filepath.Base(output))

	return
}
