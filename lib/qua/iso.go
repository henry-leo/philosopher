package qua

import (
	"fmt"
	"math"
	"strings"

	"github.com/nesvilab/philosopher/lib/mzn"
	"github.com/nesvilab/philosopher/lib/rep"
	"github.com/nesvilab/philosopher/lib/tmt"
	"github.com/nesvilab/philosopher/lib/uti"
)

const (
	mzDeltaWindow float64 = 0.5
)

// calculateIonPurity verifies how much interference there is on the precursor scans for each fragment
func calculateIonPurity(d, f string, mz mzn.MsData, evi []rep.PSMEvidence) []rep.PSMEvidence {

	// index MS1 and MS2 spectra in a dictionary
	var indexedMS1 = make(map[string]mzn.Spectrum)
	var indexedMS2 = make(map[string]mzn.Spectrum)

	for i := range mz.Spectra {

		if mz.Spectra[i].Level == "1" {

			// left-pad the spectrum index
			paddedIndex := fmt.Sprintf("%05s", mz.Spectra[i].Index)

			// left-pad the spectrum scan
			paddedScan := fmt.Sprintf("%05s", mz.Spectra[i].Scan)

			mz.Spectra[i].Index = paddedIndex
			mz.Spectra[i].Scan = paddedScan

			indexedMS1[paddedScan] = mz.Spectra[i]

		} else if mz.Spectra[i].Level == "2" {

			if mz.Spectra[i].Precursor.IsolationWindowLowerOffset == 0 && mz.Spectra[i].Precursor.IsolationWindowUpperOffset == 0 {
				mz.Spectra[i].Precursor.IsolationWindowLowerOffset = mzDeltaWindow
				mz.Spectra[i].Precursor.IsolationWindowUpperOffset = mzDeltaWindow
			}

			// left-pad the spectrum index
			paddedIndex := fmt.Sprintf("%05s", mz.Spectra[i].Index)

			// left-pad the spectrum scan
			paddedScan := fmt.Sprintf("%05s", mz.Spectra[i].Scan)

			// left-pad the precursor spectrum index
			paddedPI := fmt.Sprintf("%05s", mz.Spectra[i].Precursor.ParentIndex)

			// left-pad the precursor spectrum scan
			paddedPS := fmt.Sprintf("%05s", mz.Spectra[i].Precursor.ParentScan)

			mz.Spectra[i].Index = paddedIndex
			mz.Spectra[i].Scan = paddedScan
			mz.Spectra[i].Precursor.ParentIndex = paddedPI
			mz.Spectra[i].Precursor.ParentScan = paddedPS

			indexedMS2[paddedScan] = mz.Spectra[i]

		}

	}

	for i := range evi {

		// get spectrum index
		split := strings.Split(evi[i].Spectrum, ".")

		v2, ok := indexedMS2[split[1]]
		if ok {

			v1 := indexedMS1[v2.Precursor.ParentScan]

			var ions = make(map[float64]float64)
			var isolationWindowSummedInt float64

			for k := range v1.Mz.DecodedStream {
				if v1.Mz.DecodedStream[k] >= (v2.Precursor.TargetIon-v2.Precursor.IsolationWindowUpperOffset) && v1.Mz.DecodedStream[k] <= (v2.Precursor.TargetIon+v2.Precursor.IsolationWindowUpperOffset) {
					ions[v1.Mz.DecodedStream[k]] = v1.Intensity.DecodedStream[k]
					isolationWindowSummedInt += v1.Intensity.DecodedStream[k]
				}
			}

			// create the list of mz differences for each peak
			var mzRatio []float64
			for k := 1; k <= 6; k++ {
				r := float64(k) * (float64(1) / float64(v2.Precursor.ChargeState))
				mzRatio = append(mzRatio, uti.ToFixed(r, 2))
			}

			var isotopePackage = make(map[float64]float64)

			isotopePackage[v2.Precursor.TargetIon] = v2.Precursor.PeakIntensity
			isotopesInt := v2.Precursor.PeakIntensity

			for k, v := range ions {
				for _, m := range mzRatio {
					if math.Abs(v2.Precursor.TargetIon-k) <= (m+0.02) && math.Abs(v2.Precursor.TargetIon-k) >= (m-0.02) {
						isotopePackage[k] = v
						isotopesInt += v
						break
					}
				}
			}

			if isotopesInt == 0 {
				evi[i].Purity = 0
			} else {
				evi[i].Purity = uti.Round((isotopesInt / isolationWindowSummedInt), 5, 2)
			}

			if evi[i].Purity > 1 {
				evi[i].Purity = 1
			}

		}

	}

	return evi
}

// prepareLabelStructureWithMS2 instantiates the Label objects and maps them against the fragment scans in order to get the channel intensities
func prepareLabelStructureWithMS2(dir, format, plex string, tol float64, mz mzn.MsData) map[string]tmt.Labels {

	// get all spectra names from PSMs and create the label list
	var labels = make(map[string]tmt.Labels)
	ppmPrecision := tol / math.Pow(10, 6)

	for _, i := range mz.Spectra {
		if i.Level == "2" {

			tmt := tmt.New(plex)

			// left-pad the spectrum scan
			paddedScan := fmt.Sprintf("%05s", i.Scan)

			tmt.Index = i.Index
			tmt.Scan = paddedScan
			tmt.ChargeState = i.Precursor.ChargeState

			for j := range i.Mz.DecodedStream {

				if i.Mz.DecodedStream[j] <= (tmt.Channel1.Mz+(ppmPrecision*tmt.Channel1.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel1.Mz-(ppmPrecision*tmt.Channel1.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel1.Intensity {
						tmt.Channel1.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel2.Mz+(ppmPrecision*tmt.Channel2.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel2.Mz-(ppmPrecision*tmt.Channel2.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel2.Intensity {
						tmt.Channel2.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel3.Mz+(ppmPrecision*tmt.Channel3.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel3.Mz-(ppmPrecision*tmt.Channel3.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel3.Intensity {
						tmt.Channel3.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel4.Mz+(ppmPrecision*tmt.Channel4.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel4.Mz-(ppmPrecision*tmt.Channel4.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel4.Intensity {
						tmt.Channel4.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel5.Mz+(ppmPrecision*tmt.Channel5.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel5.Mz-(ppmPrecision*tmt.Channel5.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel5.Intensity {
						tmt.Channel5.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel6.Mz+(ppmPrecision*tmt.Channel6.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel6.Mz-(ppmPrecision*tmt.Channel6.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel6.Intensity {
						tmt.Channel6.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel7.Mz+(ppmPrecision*tmt.Channel7.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel7.Mz-(ppmPrecision*tmt.Channel7.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel7.Intensity {
						tmt.Channel7.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel8.Mz+(ppmPrecision*tmt.Channel8.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel8.Mz-(ppmPrecision*tmt.Channel8.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel8.Intensity {
						tmt.Channel8.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel9.Mz+(ppmPrecision*tmt.Channel9.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel9.Mz-(ppmPrecision*tmt.Channel9.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel9.Intensity {
						tmt.Channel9.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel10.Mz+(ppmPrecision*tmt.Channel10.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel10.Mz-(ppmPrecision*tmt.Channel10.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel10.Intensity {
						tmt.Channel10.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel11.Mz+(ppmPrecision*tmt.Channel11.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel11.Mz-(ppmPrecision*tmt.Channel11.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel11.Intensity {
						tmt.Channel11.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel12.Mz+(ppmPrecision*tmt.Channel12.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel12.Mz-(ppmPrecision*tmt.Channel12.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel12.Intensity {
						tmt.Channel12.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel13.Mz+(ppmPrecision*tmt.Channel13.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel13.Mz-(ppmPrecision*tmt.Channel13.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel13.Intensity {
						tmt.Channel13.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel14.Mz+(ppmPrecision*tmt.Channel14.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel14.Mz-(ppmPrecision*tmt.Channel14.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel14.Intensity {
						tmt.Channel14.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel15.Mz+(ppmPrecision*tmt.Channel15.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel15.Mz-(ppmPrecision*tmt.Channel15.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel15.Intensity {
						tmt.Channel15.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel16.Mz+(ppmPrecision*tmt.Channel16.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel16.Mz-(ppmPrecision*tmt.Channel16.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel16.Intensity {
						tmt.Channel16.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] > 135 {
					break
				}

			}

			labels[paddedScan] = tmt

		}
	}

	return labels
}

// prepareLabelStructureWithMS3 instantiates the Label objects and maps them against the fragment scans in order to get the channel intensities
func prepareLabelStructureWithMS3(dir, format, plex string, tol float64, mz mzn.MsData) map[string]tmt.Labels {

	// get all spectra names from PSMs and create the label list
	var labels = make(map[string]tmt.Labels)
	ppmPrecision := tol / math.Pow(10, 6)

	for _, i := range mz.Spectra {
		if i.Level == "3" {

			tmt := tmt.New(plex)

			// left-pad the spectrum scan
			paddedScan := fmt.Sprintf("%05s", i.Scan)
			precPaddedScan := fmt.Sprintf("%05s", i.Precursor.ParentScan)

			tmt.Index = i.Index
			tmt.Scan = paddedScan
			tmt.ChargeState = i.Precursor.ChargeState

			for j := range i.Mz.DecodedStream {

				if i.Mz.DecodedStream[j] <= (tmt.Channel1.Mz+(ppmPrecision*tmt.Channel1.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel1.Mz-(ppmPrecision*tmt.Channel1.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel1.Intensity {
						tmt.Channel1.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel2.Mz+(ppmPrecision*tmt.Channel2.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel2.Mz-(ppmPrecision*tmt.Channel2.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel2.Intensity {
						tmt.Channel2.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel3.Mz+(ppmPrecision*tmt.Channel3.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel3.Mz-(ppmPrecision*tmt.Channel3.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel3.Intensity {
						tmt.Channel3.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel4.Mz+(ppmPrecision*tmt.Channel4.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel4.Mz-(ppmPrecision*tmt.Channel4.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel4.Intensity {
						tmt.Channel4.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel5.Mz+(ppmPrecision*tmt.Channel5.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel5.Mz-(ppmPrecision*tmt.Channel5.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel5.Intensity {
						tmt.Channel5.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel6.Mz+(ppmPrecision*tmt.Channel6.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel6.Mz-(ppmPrecision*tmt.Channel6.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel6.Intensity {
						tmt.Channel6.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel7.Mz+(ppmPrecision*tmt.Channel7.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel7.Mz-(ppmPrecision*tmt.Channel7.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel7.Intensity {
						tmt.Channel7.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel8.Mz+(ppmPrecision*tmt.Channel8.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel8.Mz-(ppmPrecision*tmt.Channel8.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel8.Intensity {
						tmt.Channel8.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel9.Mz+(ppmPrecision*tmt.Channel9.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel9.Mz-(ppmPrecision*tmt.Channel9.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel9.Intensity {
						tmt.Channel9.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel10.Mz+(ppmPrecision*tmt.Channel10.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel10.Mz-(ppmPrecision*tmt.Channel10.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel10.Intensity {
						tmt.Channel10.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel11.Mz+(ppmPrecision*tmt.Channel11.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel11.Mz-(ppmPrecision*tmt.Channel11.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel11.Intensity {
						tmt.Channel11.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel12.Mz+(ppmPrecision*tmt.Channel12.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel12.Mz-(ppmPrecision*tmt.Channel12.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel12.Intensity {
						tmt.Channel12.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel13.Mz+(ppmPrecision*tmt.Channel13.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel13.Mz-(ppmPrecision*tmt.Channel13.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel13.Intensity {
						tmt.Channel13.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel14.Mz+(ppmPrecision*tmt.Channel14.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel14.Mz-(ppmPrecision*tmt.Channel14.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel14.Intensity {
						tmt.Channel14.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel15.Mz+(ppmPrecision*tmt.Channel15.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel15.Mz-(ppmPrecision*tmt.Channel15.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel15.Intensity {
						tmt.Channel15.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] <= (tmt.Channel16.Mz+(ppmPrecision*tmt.Channel16.Mz)) && i.Mz.DecodedStream[j] >= (tmt.Channel16.Mz-(ppmPrecision*tmt.Channel16.Mz)) {
					if i.Intensity.DecodedStream[j] > tmt.Channel16.Intensity {
						tmt.Channel16.Intensity = i.Intensity.DecodedStream[j]
					}
				}

				if i.Mz.DecodedStream[j] > 135 {
					break
				}

			}

			labels[precPaddedScan] = tmt

		}
	}

	return labels
}

// mapLabeledSpectra maps all labeled spectra to PSMs
func mapLabeledSpectra(labels map[string]tmt.Labels, purity float64, evi []rep.PSMEvidence) []rep.PSMEvidence {

	for i := range evi {

		split := strings.Split(evi[i].Spectrum, ".")

		// referenced by scan number
		v, ok := labels[split[2]]
		if ok {

			evi[i].Labels.Spectrum = v.Spectrum
			evi[i].Labels.Index = v.Index
			evi[i].Labels.Scan = v.Scan
			evi[i].Labels.Channel1.Intensity = v.Channel1.Intensity
			evi[i].Labels.Channel2.Intensity = v.Channel2.Intensity
			evi[i].Labels.Channel3.Intensity = v.Channel3.Intensity
			evi[i].Labels.Channel4.Intensity = v.Channel4.Intensity
			evi[i].Labels.Channel5.Intensity = v.Channel5.Intensity
			evi[i].Labels.Channel6.Intensity = v.Channel6.Intensity
			evi[i].Labels.Channel7.Intensity = v.Channel7.Intensity
			evi[i].Labels.Channel8.Intensity = v.Channel8.Intensity
			evi[i].Labels.Channel9.Intensity = v.Channel9.Intensity
			evi[i].Labels.Channel10.Intensity = v.Channel10.Intensity
			evi[i].Labels.Channel11.Intensity = v.Channel11.Intensity
			evi[i].Labels.Channel12.Intensity = v.Channel12.Intensity
			evi[i].Labels.Channel13.Intensity = v.Channel13.Intensity
			evi[i].Labels.Channel14.Intensity = v.Channel14.Intensity
			evi[i].Labels.Channel15.Intensity = v.Channel15.Intensity
			evi[i].Labels.Channel16.Intensity = v.Channel16.Intensity

		}
	}

	return evi
}

// the assignment of usage is only done for general PSM, not for phosphoPSMs
func assignUsage(evi rep.Evidence, spectrumMap map[string]tmt.Labels) rep.Evidence {

	for i := range evi.PSM {
		_, ok := spectrumMap[evi.PSM[i].Spectrum]
		if ok {
			evi.PSM[i].Labels.IsUsed = true
		}
	}

	return evi
}

func correctUnlabelledSpectra(evi rep.Evidence) rep.Evidence {

	for i := range evi.PSM {

		var flag = 0

		if len(evi.PSM[i].Modifications.Index) < 1 {
			evi.PSM[i].Labels.Channel1.Intensity = 0
			evi.PSM[i].Labels.Channel2.Intensity = 0
			evi.PSM[i].Labels.Channel3.Intensity = 0
			evi.PSM[i].Labels.Channel4.Intensity = 0
			evi.PSM[i].Labels.Channel5.Intensity = 0
			evi.PSM[i].Labels.Channel6.Intensity = 0
			evi.PSM[i].Labels.Channel7.Intensity = 0
			evi.PSM[i].Labels.Channel8.Intensity = 0
			evi.PSM[i].Labels.Channel9.Intensity = 0
			evi.PSM[i].Labels.Channel10.Intensity = 0
			evi.PSM[i].Labels.Channel11.Intensity = 0
			evi.PSM[i].Labels.Channel12.Intensity = 0
			evi.PSM[i].Labels.Channel13.Intensity = 0
			evi.PSM[i].Labels.Channel14.Intensity = 0
			evi.PSM[i].Labels.Channel15.Intensity = 0
			evi.PSM[i].Labels.Channel16.Intensity = 0
		} else {
			for _, j := range evi.PSM[i].Modifications.Index {
				if j.MassDiff >= 229.1629 {
					flag++
				}
			}

			if flag == 0 {
				evi.PSM[i].Labels.Channel1.Intensity = 0
				evi.PSM[i].Labels.Channel2.Intensity = 0
				evi.PSM[i].Labels.Channel3.Intensity = 0
				evi.PSM[i].Labels.Channel4.Intensity = 0
				evi.PSM[i].Labels.Channel5.Intensity = 0
				evi.PSM[i].Labels.Channel6.Intensity = 0
				evi.PSM[i].Labels.Channel7.Intensity = 0
				evi.PSM[i].Labels.Channel8.Intensity = 0
				evi.PSM[i].Labels.Channel9.Intensity = 0
				evi.PSM[i].Labels.Channel10.Intensity = 0
				evi.PSM[i].Labels.Channel11.Intensity = 0
				evi.PSM[i].Labels.Channel12.Intensity = 0
				evi.PSM[i].Labels.Channel13.Intensity = 0
				evi.PSM[i].Labels.Channel14.Intensity = 0
				evi.PSM[i].Labels.Channel15.Intensity = 0
				evi.PSM[i].Labels.Channel16.Intensity = 0
			}

		}
	}

	return evi
}

// rollUpPeptides gathers PSM info and filters them before summing the instensities to the peptide level
func rollUpPeptides(evi rep.Evidence, spectrumMap map[string]tmt.Labels, phosphoSpectrumMap map[string]tmt.Labels) rep.Evidence {

	for j := range evi.Peptides {
		for k := range evi.Peptides[j].Spectra {

			i, ok := spectrumMap[k]
			if ok {

				evi.Peptides[j].Labels.Channel1.Name = i.Channel1.Name
				evi.Peptides[j].Labels.Channel1.Mz = i.Channel1.Mz
				evi.Peptides[j].Labels.Channel1.Intensity += i.Channel1.Intensity

				evi.Peptides[j].Labels.Channel2.Name = i.Channel2.Name
				evi.Peptides[j].Labels.Channel2.Mz = i.Channel2.Mz
				evi.Peptides[j].Labels.Channel2.Intensity += i.Channel2.Intensity

				evi.Peptides[j].Labels.Channel3.Name = i.Channel3.Name
				evi.Peptides[j].Labels.Channel3.Mz = i.Channel3.Mz
				evi.Peptides[j].Labels.Channel3.Intensity += i.Channel3.Intensity

				evi.Peptides[j].Labels.Channel4.Name = i.Channel4.Name
				evi.Peptides[j].Labels.Channel4.Mz = i.Channel4.Mz
				evi.Peptides[j].Labels.Channel4.Intensity += i.Channel4.Intensity

				evi.Peptides[j].Labels.Channel5.Name = i.Channel5.Name
				evi.Peptides[j].Labels.Channel5.Mz = i.Channel5.Mz
				evi.Peptides[j].Labels.Channel5.Intensity += i.Channel5.Intensity

				evi.Peptides[j].Labels.Channel6.Name = i.Channel6.Name
				evi.Peptides[j].Labels.Channel6.Mz = i.Channel6.Mz
				evi.Peptides[j].Labels.Channel6.Intensity += i.Channel6.Intensity

				evi.Peptides[j].Labels.Channel7.Name = i.Channel7.Name
				evi.Peptides[j].Labels.Channel7.Mz = i.Channel7.Mz
				evi.Peptides[j].Labels.Channel7.Intensity += i.Channel7.Intensity

				evi.Peptides[j].Labels.Channel8.Name = i.Channel8.Name
				evi.Peptides[j].Labels.Channel8.Mz = i.Channel8.Mz
				evi.Peptides[j].Labels.Channel8.Intensity += i.Channel8.Intensity

				evi.Peptides[j].Labels.Channel9.Name = i.Channel9.Name
				evi.Peptides[j].Labels.Channel9.Mz = i.Channel9.Mz
				evi.Peptides[j].Labels.Channel9.Intensity += i.Channel9.Intensity

				evi.Peptides[j].Labels.Channel10.Name = i.Channel10.Name
				evi.Peptides[j].Labels.Channel10.Mz = i.Channel10.Mz
				evi.Peptides[j].Labels.Channel10.Intensity += i.Channel10.Intensity

				evi.Peptides[j].Labels.Channel11.Name = i.Channel11.Name
				evi.Peptides[j].Labels.Channel11.Mz = i.Channel11.Mz
				evi.Peptides[j].Labels.Channel11.Intensity += i.Channel11.Intensity

				evi.Peptides[j].Labels.Channel12.Name = i.Channel12.Name
				evi.Peptides[j].Labels.Channel12.Mz = i.Channel12.Mz
				evi.Peptides[j].Labels.Channel12.Intensity += i.Channel12.Intensity

				evi.Peptides[j].Labels.Channel13.Name = i.Channel13.Name
				evi.Peptides[j].Labels.Channel13.Mz = i.Channel13.Mz
				evi.Peptides[j].Labels.Channel13.Intensity += i.Channel13.Intensity

				evi.Peptides[j].Labels.Channel14.Name = i.Channel14.Name
				evi.Peptides[j].Labels.Channel14.Mz = i.Channel14.Mz
				evi.Peptides[j].Labels.Channel14.Intensity += i.Channel14.Intensity

				evi.Peptides[j].Labels.Channel15.Name = i.Channel15.Name
				evi.Peptides[j].Labels.Channel15.Mz = i.Channel15.Mz
				evi.Peptides[j].Labels.Channel15.Intensity += i.Channel15.Intensity

				evi.Peptides[j].Labels.Channel16.Name = i.Channel16.Name
				evi.Peptides[j].Labels.Channel16.Mz = i.Channel16.Mz
				evi.Peptides[j].Labels.Channel16.Intensity += i.Channel16.Intensity
			}

			i, ok = phosphoSpectrumMap[k]
			if ok {

				evi.Peptides[j].PhosphoLabels.Channel1.Name = i.Channel1.Name
				evi.Peptides[j].PhosphoLabels.Channel1.Mz = i.Channel1.Mz
				evi.Peptides[j].PhosphoLabels.Channel1.Intensity += i.Channel1.Intensity

				evi.Peptides[j].PhosphoLabels.Channel2.Name = i.Channel2.Name
				evi.Peptides[j].PhosphoLabels.Channel2.Mz = i.Channel2.Mz
				evi.Peptides[j].PhosphoLabels.Channel2.Intensity += i.Channel2.Intensity

				evi.Peptides[j].PhosphoLabels.Channel3.Name = i.Channel3.Name
				evi.Peptides[j].PhosphoLabels.Channel3.Mz = i.Channel3.Mz
				evi.Peptides[j].PhosphoLabels.Channel3.Intensity += i.Channel3.Intensity

				evi.Peptides[j].PhosphoLabels.Channel4.Name = i.Channel4.Name
				evi.Peptides[j].PhosphoLabels.Channel4.Mz = i.Channel4.Mz
				evi.Peptides[j].PhosphoLabels.Channel4.Intensity += i.Channel4.Intensity

				evi.Peptides[j].PhosphoLabels.Channel5.Name = i.Channel5.Name
				evi.Peptides[j].PhosphoLabels.Channel5.Mz = i.Channel5.Mz
				evi.Peptides[j].PhosphoLabels.Channel5.Intensity += i.Channel5.Intensity

				evi.Peptides[j].PhosphoLabels.Channel6.Name = i.Channel6.Name
				evi.Peptides[j].PhosphoLabels.Channel6.Mz = i.Channel6.Mz
				evi.Peptides[j].PhosphoLabels.Channel6.Intensity += i.Channel6.Intensity

				evi.Peptides[j].PhosphoLabels.Channel7.Name = i.Channel7.Name
				evi.Peptides[j].PhosphoLabels.Channel7.Mz = i.Channel7.Mz
				evi.Peptides[j].PhosphoLabels.Channel7.Intensity += i.Channel7.Intensity

				evi.Peptides[j].PhosphoLabels.Channel8.Name = i.Channel8.Name
				evi.Peptides[j].PhosphoLabels.Channel8.Mz = i.Channel8.Mz
				evi.Peptides[j].PhosphoLabels.Channel8.Intensity += i.Channel8.Intensity

				evi.Peptides[j].PhosphoLabels.Channel9.Name = i.Channel9.Name
				evi.Peptides[j].PhosphoLabels.Channel9.Mz = i.Channel9.Mz
				evi.Peptides[j].PhosphoLabels.Channel9.Intensity += i.Channel9.Intensity

				evi.Peptides[j].PhosphoLabels.Channel10.Name = i.Channel10.Name
				evi.Peptides[j].PhosphoLabels.Channel10.Mz = i.Channel10.Mz
				evi.Peptides[j].PhosphoLabels.Channel10.Intensity += i.Channel10.Intensity

				evi.Peptides[j].PhosphoLabels.Channel11.Name = i.Channel11.Name
				evi.Peptides[j].PhosphoLabels.Channel11.Mz = i.Channel11.Mz
				evi.Peptides[j].PhosphoLabels.Channel11.Intensity += i.Channel11.Intensity

				evi.Peptides[j].PhosphoLabels.Channel12.Name = i.Channel12.Name
				evi.Peptides[j].PhosphoLabels.Channel12.Mz = i.Channel12.Mz
				evi.Peptides[j].PhosphoLabels.Channel12.Intensity += i.Channel12.Intensity

				evi.Peptides[j].PhosphoLabels.Channel13.Name = i.Channel13.Name
				evi.Peptides[j].PhosphoLabels.Channel13.Mz = i.Channel13.Mz
				evi.Peptides[j].PhosphoLabels.Channel13.Intensity += i.Channel13.Intensity

				evi.Peptides[j].PhosphoLabels.Channel14.Name = i.Channel14.Name
				evi.Peptides[j].PhosphoLabels.Channel14.Mz = i.Channel14.Mz
				evi.Peptides[j].PhosphoLabels.Channel14.Intensity += i.Channel14.Intensity

				evi.Peptides[j].PhosphoLabels.Channel15.Name = i.Channel15.Name
				evi.Peptides[j].PhosphoLabels.Channel15.Mz = i.Channel15.Mz
				evi.Peptides[j].PhosphoLabels.Channel15.Intensity += i.Channel15.Intensity

				evi.Peptides[j].PhosphoLabels.Channel16.Name = i.Channel16.Name
				evi.Peptides[j].PhosphoLabels.Channel16.Mz = i.Channel16.Mz
				evi.Peptides[j].PhosphoLabels.Channel16.Intensity += i.Channel16.Intensity
			}

		}
	}

	return evi
}

// rollUpPeptideIons gathers PSM info and filters them before summing the instensities to the peptide ION level
func rollUpPeptideIons(evi rep.Evidence, spectrumMap map[string]tmt.Labels, phosphoSpectrumMap map[string]tmt.Labels) rep.Evidence {

	for j := range evi.Ions {
		for k := range evi.Ions[j].Spectra {

			i, ok := spectrumMap[k]
			if ok {

				evi.Ions[j].Labels.Channel1.Name = i.Channel1.Name
				evi.Ions[j].Labels.Channel1.Mz = i.Channel1.Mz
				evi.Ions[j].Labels.Channel1.Intensity += i.Channel1.Intensity

				evi.Ions[j].Labels.Channel2.Name = i.Channel2.Name
				evi.Ions[j].Labels.Channel2.Mz = i.Channel2.Mz
				evi.Ions[j].Labels.Channel2.Intensity += i.Channel2.Intensity

				evi.Ions[j].Labels.Channel3.Name = i.Channel3.Name
				evi.Ions[j].Labels.Channel3.Mz = i.Channel3.Mz
				evi.Ions[j].Labels.Channel3.Intensity += i.Channel3.Intensity

				evi.Ions[j].Labels.Channel4.Name = i.Channel4.Name
				evi.Ions[j].Labels.Channel4.Mz = i.Channel4.Mz
				evi.Ions[j].Labels.Channel4.Intensity += i.Channel4.Intensity

				evi.Ions[j].Labels.Channel5.Name = i.Channel5.Name
				evi.Ions[j].Labels.Channel5.Mz = i.Channel5.Mz
				evi.Ions[j].Labels.Channel5.Intensity += i.Channel5.Intensity

				evi.Ions[j].Labels.Channel6.Name = i.Channel6.Name
				evi.Ions[j].Labels.Channel6.Mz = i.Channel6.Mz
				evi.Ions[j].Labels.Channel6.Intensity += i.Channel6.Intensity

				evi.Ions[j].Labels.Channel7.Name = i.Channel7.Name
				evi.Ions[j].Labels.Channel7.Mz = i.Channel7.Mz
				evi.Ions[j].Labels.Channel7.Intensity += i.Channel7.Intensity

				evi.Ions[j].Labels.Channel8.Name = i.Channel8.Name
				evi.Ions[j].Labels.Channel8.Mz = i.Channel8.Mz
				evi.Ions[j].Labels.Channel8.Intensity += i.Channel8.Intensity

				evi.Ions[j].Labels.Channel9.Name = i.Channel9.Name
				evi.Ions[j].Labels.Channel9.Mz = i.Channel9.Mz
				evi.Ions[j].Labels.Channel9.Intensity += i.Channel9.Intensity

				evi.Ions[j].Labels.Channel10.Name = i.Channel10.Name
				evi.Ions[j].Labels.Channel10.Mz = i.Channel10.Mz
				evi.Ions[j].Labels.Channel10.Intensity += i.Channel10.Intensity

				evi.Ions[j].Labels.Channel11.Name = i.Channel11.Name
				evi.Ions[j].Labels.Channel11.Mz = i.Channel11.Mz
				evi.Ions[j].Labels.Channel11.Intensity += i.Channel11.Intensity

				evi.Ions[j].Labels.Channel12.Name = i.Channel12.Name
				evi.Ions[j].Labels.Channel12.Mz = i.Channel12.Mz
				evi.Ions[j].Labels.Channel12.Intensity += i.Channel12.Intensity

				evi.Ions[j].Labels.Channel13.Name = i.Channel13.Name
				evi.Ions[j].Labels.Channel13.Mz = i.Channel13.Mz
				evi.Ions[j].Labels.Channel13.Intensity += i.Channel13.Intensity

				evi.Ions[j].Labels.Channel14.Name = i.Channel14.Name
				evi.Ions[j].Labels.Channel14.Mz = i.Channel14.Mz
				evi.Ions[j].Labels.Channel14.Intensity += i.Channel14.Intensity

				evi.Ions[j].Labels.Channel15.Name = i.Channel15.Name
				evi.Ions[j].Labels.Channel15.Mz = i.Channel15.Mz
				evi.Ions[j].Labels.Channel15.Intensity += i.Channel15.Intensity

				evi.Ions[j].Labels.Channel16.Name = i.Channel16.Name
				evi.Ions[j].Labels.Channel16.Mz = i.Channel16.Mz
				evi.Ions[j].Labels.Channel16.Intensity += i.Channel16.Intensity
			}

			i, ok = phosphoSpectrumMap[k]
			if ok {

				evi.Ions[j].PhosphoLabels.Channel1.Name = i.Channel1.Name
				evi.Ions[j].PhosphoLabels.Channel1.Mz = i.Channel1.Mz
				evi.Ions[j].PhosphoLabels.Channel1.Intensity += i.Channel1.Intensity

				evi.Ions[j].PhosphoLabels.Channel2.Name = i.Channel2.Name
				evi.Ions[j].PhosphoLabels.Channel2.Mz = i.Channel2.Mz
				evi.Ions[j].PhosphoLabels.Channel2.Intensity += i.Channel2.Intensity

				evi.Ions[j].PhosphoLabels.Channel3.Name = i.Channel3.Name
				evi.Ions[j].PhosphoLabels.Channel3.Mz = i.Channel3.Mz
				evi.Ions[j].PhosphoLabels.Channel3.Intensity += i.Channel3.Intensity

				evi.Ions[j].PhosphoLabels.Channel4.Name = i.Channel4.Name
				evi.Ions[j].PhosphoLabels.Channel4.Mz = i.Channel4.Mz
				evi.Ions[j].PhosphoLabels.Channel4.Intensity += i.Channel4.Intensity

				evi.Ions[j].PhosphoLabels.Channel5.Name = i.Channel5.Name
				evi.Ions[j].PhosphoLabels.Channel5.Mz = i.Channel5.Mz
				evi.Ions[j].PhosphoLabels.Channel5.Intensity += i.Channel5.Intensity

				evi.Ions[j].PhosphoLabels.Channel6.Name = i.Channel6.Name
				evi.Ions[j].PhosphoLabels.Channel6.Mz = i.Channel6.Mz
				evi.Ions[j].PhosphoLabels.Channel6.Intensity += i.Channel6.Intensity

				evi.Ions[j].PhosphoLabels.Channel7.Name = i.Channel7.Name
				evi.Ions[j].PhosphoLabels.Channel7.Mz = i.Channel7.Mz
				evi.Ions[j].PhosphoLabels.Channel7.Intensity += i.Channel7.Intensity

				evi.Ions[j].PhosphoLabels.Channel8.Name = i.Channel8.Name
				evi.Ions[j].PhosphoLabels.Channel8.Mz = i.Channel8.Mz
				evi.Ions[j].PhosphoLabels.Channel8.Intensity += i.Channel8.Intensity

				evi.Ions[j].PhosphoLabels.Channel9.Name = i.Channel9.Name
				evi.Ions[j].PhosphoLabels.Channel9.Mz = i.Channel9.Mz
				evi.Ions[j].PhosphoLabels.Channel9.Intensity += i.Channel9.Intensity

				evi.Ions[j].PhosphoLabels.Channel10.Name = i.Channel10.Name
				evi.Ions[j].PhosphoLabels.Channel10.Mz = i.Channel10.Mz
				evi.Ions[j].PhosphoLabels.Channel10.Intensity += i.Channel10.Intensity

				evi.Ions[j].PhosphoLabels.Channel11.Name = i.Channel11.Name
				evi.Ions[j].PhosphoLabels.Channel11.Mz = i.Channel11.Mz
				evi.Ions[j].PhosphoLabels.Channel11.Intensity += i.Channel11.Intensity

				evi.Ions[j].PhosphoLabels.Channel12.Name = i.Channel12.Name
				evi.Ions[j].PhosphoLabels.Channel12.Mz = i.Channel12.Mz
				evi.Ions[j].PhosphoLabels.Channel12.Intensity += i.Channel12.Intensity

				evi.Ions[j].PhosphoLabels.Channel13.Name = i.Channel13.Name
				evi.Ions[j].PhosphoLabels.Channel13.Mz = i.Channel13.Mz
				evi.Ions[j].PhosphoLabels.Channel13.Intensity += i.Channel13.Intensity

				evi.Ions[j].PhosphoLabels.Channel14.Name = i.Channel14.Name
				evi.Ions[j].PhosphoLabels.Channel14.Mz = i.Channel14.Mz
				evi.Ions[j].PhosphoLabels.Channel14.Intensity += i.Channel14.Intensity

				evi.Ions[j].PhosphoLabels.Channel15.Name = i.Channel15.Name
				evi.Ions[j].PhosphoLabels.Channel15.Mz = i.Channel15.Mz
				evi.Ions[j].PhosphoLabels.Channel15.Intensity += i.Channel15.Intensity

				evi.Ions[j].PhosphoLabels.Channel16.Name = i.Channel16.Name
				evi.Ions[j].PhosphoLabels.Channel16.Mz = i.Channel16.Mz
				evi.Ions[j].PhosphoLabels.Channel16.Intensity += i.Channel16.Intensity
			}

		}
	}

	return evi
}

// rollUpProteins gathers PSM info and filters them before summing the instensities to the peptide ION level
func rollUpProteins(evi rep.Evidence, spectrumMap map[string]tmt.Labels, phosphoSpectrumMap map[string]tmt.Labels) rep.Evidence {

	for j := range evi.Proteins {
		for _, k := range evi.Proteins[j].TotalPeptideIons {
			for l := range k.Spectra {

				i, ok := spectrumMap[l]
				if ok {
					evi.Proteins[j].TotalLabels.Channel1.Name = i.Channel1.Name
					evi.Proteins[j].TotalLabels.Channel1.Mz = i.Channel1.Mz
					evi.Proteins[j].TotalLabels.Channel1.Intensity += i.Channel1.Intensity

					evi.Proteins[j].TotalLabels.Channel2.Name = i.Channel2.Name
					evi.Proteins[j].TotalLabels.Channel2.Mz = i.Channel2.Mz
					evi.Proteins[j].TotalLabels.Channel2.Intensity += i.Channel2.Intensity

					evi.Proteins[j].TotalLabels.Channel3.Name = i.Channel3.Name
					evi.Proteins[j].TotalLabels.Channel3.Mz = i.Channel3.Mz
					evi.Proteins[j].TotalLabels.Channel3.Intensity += i.Channel3.Intensity

					evi.Proteins[j].TotalLabels.Channel4.Name = i.Channel4.Name
					evi.Proteins[j].TotalLabels.Channel4.Mz = i.Channel4.Mz
					evi.Proteins[j].TotalLabels.Channel4.Intensity += i.Channel4.Intensity

					evi.Proteins[j].TotalLabels.Channel5.Name = i.Channel5.Name
					evi.Proteins[j].TotalLabels.Channel5.Mz = i.Channel5.Mz
					evi.Proteins[j].TotalLabels.Channel5.Intensity += i.Channel5.Intensity

					evi.Proteins[j].TotalLabels.Channel6.Name = i.Channel6.Name
					evi.Proteins[j].TotalLabels.Channel6.Mz = i.Channel6.Mz
					evi.Proteins[j].TotalLabels.Channel6.Intensity += i.Channel6.Intensity

					evi.Proteins[j].TotalLabels.Channel7.Name = i.Channel7.Name
					evi.Proteins[j].TotalLabels.Channel7.Mz = i.Channel7.Mz
					evi.Proteins[j].TotalLabels.Channel7.Intensity += i.Channel7.Intensity

					evi.Proteins[j].TotalLabels.Channel8.Name = i.Channel8.Name
					evi.Proteins[j].TotalLabels.Channel8.Mz = i.Channel8.Mz
					evi.Proteins[j].TotalLabels.Channel8.Intensity += i.Channel8.Intensity

					evi.Proteins[j].TotalLabels.Channel9.Name = i.Channel9.Name
					evi.Proteins[j].TotalLabels.Channel9.Mz = i.Channel9.Mz
					evi.Proteins[j].TotalLabels.Channel9.Intensity += i.Channel9.Intensity

					evi.Proteins[j].TotalLabels.Channel10.Name = i.Channel10.Name
					evi.Proteins[j].TotalLabels.Channel10.Mz = i.Channel10.Mz
					evi.Proteins[j].TotalLabels.Channel10.Intensity += i.Channel10.Intensity

					evi.Proteins[j].TotalLabels.Channel11.Name = i.Channel11.Name
					evi.Proteins[j].TotalLabels.Channel11.Mz = i.Channel11.Mz
					evi.Proteins[j].TotalLabels.Channel11.Intensity += i.Channel11.Intensity

					evi.Proteins[j].TotalLabels.Channel12.Name = i.Channel12.Name
					evi.Proteins[j].TotalLabels.Channel12.Mz = i.Channel12.Mz
					evi.Proteins[j].TotalLabels.Channel12.Intensity += i.Channel12.Intensity

					evi.Proteins[j].TotalLabels.Channel13.Name = i.Channel13.Name
					evi.Proteins[j].TotalLabels.Channel13.Mz = i.Channel13.Mz
					evi.Proteins[j].TotalLabels.Channel13.Intensity += i.Channel13.Intensity

					evi.Proteins[j].TotalLabels.Channel14.Name = i.Channel14.Name
					evi.Proteins[j].TotalLabels.Channel14.Mz = i.Channel14.Mz
					evi.Proteins[j].TotalLabels.Channel14.Intensity += i.Channel14.Intensity

					evi.Proteins[j].TotalLabels.Channel15.Name = i.Channel15.Name
					evi.Proteins[j].TotalLabels.Channel15.Mz = i.Channel15.Mz
					evi.Proteins[j].TotalLabels.Channel15.Intensity += i.Channel15.Intensity

					evi.Proteins[j].TotalLabels.Channel16.Name = i.Channel16.Name
					evi.Proteins[j].TotalLabels.Channel16.Mz = i.Channel16.Mz
					evi.Proteins[j].TotalLabels.Channel16.Intensity += i.Channel16.Intensity

					//if k.IsNondegenerateEvidence {
					if k.IsUnique {
						evi.Proteins[j].UniqueLabels.Channel1.Name = i.Channel1.Name
						evi.Proteins[j].UniqueLabels.Channel1.Mz = i.Channel1.Mz
						evi.Proteins[j].UniqueLabels.Channel1.Intensity += i.Channel1.Intensity

						evi.Proteins[j].UniqueLabels.Channel2.Name = i.Channel2.Name
						evi.Proteins[j].UniqueLabels.Channel2.Mz = i.Channel2.Mz
						evi.Proteins[j].UniqueLabels.Channel2.Intensity += i.Channel2.Intensity

						evi.Proteins[j].UniqueLabels.Channel3.Name = i.Channel3.Name
						evi.Proteins[j].UniqueLabels.Channel3.Mz = i.Channel3.Mz
						evi.Proteins[j].UniqueLabels.Channel3.Intensity += i.Channel3.Intensity

						evi.Proteins[j].UniqueLabels.Channel4.Name = i.Channel4.Name
						evi.Proteins[j].UniqueLabels.Channel4.Mz = i.Channel4.Mz
						evi.Proteins[j].UniqueLabels.Channel4.Intensity += i.Channel4.Intensity

						evi.Proteins[j].UniqueLabels.Channel5.Name = i.Channel5.Name
						evi.Proteins[j].UniqueLabels.Channel5.Mz = i.Channel5.Mz
						evi.Proteins[j].UniqueLabels.Channel5.Intensity += i.Channel5.Intensity

						evi.Proteins[j].UniqueLabels.Channel6.Name = i.Channel6.Name
						evi.Proteins[j].UniqueLabels.Channel6.Mz = i.Channel6.Mz
						evi.Proteins[j].UniqueLabels.Channel6.Intensity += i.Channel6.Intensity

						evi.Proteins[j].UniqueLabels.Channel7.Name = i.Channel7.Name
						evi.Proteins[j].UniqueLabels.Channel7.Mz = i.Channel7.Mz
						evi.Proteins[j].UniqueLabels.Channel7.Intensity += i.Channel7.Intensity

						evi.Proteins[j].UniqueLabels.Channel8.Name = i.Channel8.Name
						evi.Proteins[j].UniqueLabels.Channel8.Mz = i.Channel8.Mz
						evi.Proteins[j].UniqueLabels.Channel8.Intensity += i.Channel8.Intensity

						evi.Proteins[j].UniqueLabels.Channel9.Name = i.Channel9.Name
						evi.Proteins[j].UniqueLabels.Channel9.Mz = i.Channel9.Mz
						evi.Proteins[j].UniqueLabels.Channel9.Intensity += i.Channel9.Intensity

						evi.Proteins[j].UniqueLabels.Channel10.Name = i.Channel10.Name
						evi.Proteins[j].UniqueLabels.Channel10.Mz = i.Channel10.Mz
						evi.Proteins[j].UniqueLabels.Channel10.Intensity += i.Channel10.Intensity

						evi.Proteins[j].UniqueLabels.Channel11.Name = i.Channel11.Name
						evi.Proteins[j].UniqueLabels.Channel11.Mz = i.Channel11.Mz
						evi.Proteins[j].UniqueLabels.Channel11.Intensity += i.Channel11.Intensity

						evi.Proteins[j].UniqueLabels.Channel12.Name = i.Channel12.Name
						evi.Proteins[j].UniqueLabels.Channel12.Mz = i.Channel12.Mz
						evi.Proteins[j].UniqueLabels.Channel12.Intensity += i.Channel12.Intensity

						evi.Proteins[j].UniqueLabels.Channel13.Name = i.Channel13.Name
						evi.Proteins[j].UniqueLabels.Channel13.Mz = i.Channel13.Mz
						evi.Proteins[j].UniqueLabels.Channel13.Intensity += i.Channel13.Intensity

						evi.Proteins[j].UniqueLabels.Channel14.Name = i.Channel14.Name
						evi.Proteins[j].UniqueLabels.Channel14.Mz = i.Channel14.Mz
						evi.Proteins[j].UniqueLabels.Channel14.Intensity += i.Channel14.Intensity

						evi.Proteins[j].UniqueLabels.Channel15.Name = i.Channel15.Name
						evi.Proteins[j].UniqueLabels.Channel15.Mz = i.Channel15.Mz
						evi.Proteins[j].UniqueLabels.Channel15.Intensity += i.Channel15.Intensity

						evi.Proteins[j].UniqueLabels.Channel16.Name = i.Channel16.Name
						evi.Proteins[j].UniqueLabels.Channel16.Mz = i.Channel16.Mz
						evi.Proteins[j].UniqueLabels.Channel16.Intensity += i.Channel16.Intensity
					}

					if k.IsURazor {
						evi.Proteins[j].URazorLabels.Channel1.Name = i.Channel1.Name
						evi.Proteins[j].URazorLabels.Channel1.Mz = i.Channel1.Mz
						evi.Proteins[j].URazorLabels.Channel1.Intensity += i.Channel1.Intensity

						evi.Proteins[j].URazorLabels.Channel2.Name = i.Channel2.Name
						evi.Proteins[j].URazorLabels.Channel2.Mz = i.Channel2.Mz
						evi.Proteins[j].URazorLabels.Channel2.Intensity += i.Channel2.Intensity

						evi.Proteins[j].URazorLabels.Channel3.Name = i.Channel3.Name
						evi.Proteins[j].URazorLabels.Channel3.Mz = i.Channel3.Mz
						evi.Proteins[j].URazorLabels.Channel3.Intensity += i.Channel3.Intensity

						evi.Proteins[j].URazorLabels.Channel4.Name = i.Channel4.Name
						evi.Proteins[j].URazorLabels.Channel4.Mz = i.Channel4.Mz
						evi.Proteins[j].URazorLabels.Channel4.Intensity += i.Channel4.Intensity

						evi.Proteins[j].URazorLabels.Channel5.Name = i.Channel5.Name
						evi.Proteins[j].URazorLabels.Channel5.Mz = i.Channel5.Mz
						evi.Proteins[j].URazorLabels.Channel5.Intensity += i.Channel5.Intensity

						evi.Proteins[j].URazorLabels.Channel6.Name = i.Channel6.Name
						evi.Proteins[j].URazorLabels.Channel6.Mz = i.Channel6.Mz
						evi.Proteins[j].URazorLabels.Channel6.Intensity += i.Channel6.Intensity

						evi.Proteins[j].URazorLabels.Channel7.Name = i.Channel7.Name
						evi.Proteins[j].URazorLabels.Channel7.Mz = i.Channel7.Mz
						evi.Proteins[j].URazorLabels.Channel7.Intensity += i.Channel7.Intensity

						evi.Proteins[j].URazorLabels.Channel8.Name = i.Channel8.Name
						evi.Proteins[j].URazorLabels.Channel8.Mz = i.Channel8.Mz
						evi.Proteins[j].URazorLabels.Channel8.Intensity += i.Channel8.Intensity

						evi.Proteins[j].URazorLabels.Channel9.Name = i.Channel9.Name
						evi.Proteins[j].URazorLabels.Channel9.Mz = i.Channel9.Mz
						evi.Proteins[j].URazorLabels.Channel9.Intensity += i.Channel9.Intensity

						evi.Proteins[j].URazorLabels.Channel10.Name = i.Channel10.Name
						evi.Proteins[j].URazorLabels.Channel10.Mz = i.Channel10.Mz
						evi.Proteins[j].URazorLabels.Channel10.Intensity += i.Channel10.Intensity

						evi.Proteins[j].URazorLabels.Channel11.Name = i.Channel11.Name
						evi.Proteins[j].URazorLabels.Channel11.Mz = i.Channel11.Mz
						evi.Proteins[j].URazorLabels.Channel11.Intensity += i.Channel11.Intensity

						evi.Proteins[j].URazorLabels.Channel12.Name = i.Channel12.Name
						evi.Proteins[j].URazorLabels.Channel12.Mz = i.Channel12.Mz
						evi.Proteins[j].URazorLabels.Channel12.Intensity += i.Channel12.Intensity

						evi.Proteins[j].URazorLabels.Channel13.Name = i.Channel13.Name
						evi.Proteins[j].URazorLabels.Channel13.Mz = i.Channel13.Mz
						evi.Proteins[j].URazorLabels.Channel13.Intensity += i.Channel13.Intensity

						evi.Proteins[j].URazorLabels.Channel14.Name = i.Channel14.Name
						evi.Proteins[j].URazorLabels.Channel14.Mz = i.Channel14.Mz
						evi.Proteins[j].URazorLabels.Channel14.Intensity += i.Channel14.Intensity

						evi.Proteins[j].URazorLabels.Channel15.Name = i.Channel15.Name
						evi.Proteins[j].URazorLabels.Channel15.Mz = i.Channel15.Mz
						evi.Proteins[j].URazorLabels.Channel15.Intensity += i.Channel15.Intensity

						evi.Proteins[j].URazorLabels.Channel16.Name = i.Channel16.Name
						evi.Proteins[j].URazorLabels.Channel16.Mz = i.Channel16.Mz
						evi.Proteins[j].URazorLabels.Channel16.Intensity += i.Channel16.Intensity
					}
				}

				i, ok = phosphoSpectrumMap[l]
				if ok {
					evi.Proteins[j].PhosphoTotalLabels.Channel1.Name = i.Channel1.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel1.Mz = i.Channel1.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel1.Intensity += i.Channel1.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel2.Name = i.Channel2.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel2.Mz = i.Channel2.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel2.Intensity += i.Channel2.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel3.Name = i.Channel3.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel3.Mz = i.Channel3.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel3.Intensity += i.Channel3.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel4.Name = i.Channel4.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel4.Mz = i.Channel4.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel4.Intensity += i.Channel4.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel5.Name = i.Channel5.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel5.Mz = i.Channel5.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel5.Intensity += i.Channel5.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel6.Name = i.Channel6.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel6.Mz = i.Channel6.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel6.Intensity += i.Channel6.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel7.Name = i.Channel7.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel7.Mz = i.Channel7.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel7.Intensity += i.Channel7.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel8.Name = i.Channel8.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel8.Mz = i.Channel8.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel8.Intensity += i.Channel8.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel9.Name = i.Channel9.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel9.Mz = i.Channel9.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel9.Intensity += i.Channel9.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel10.Name = i.Channel10.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel10.Mz = i.Channel10.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel10.Intensity += i.Channel10.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel11.Name = i.Channel11.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel11.Mz = i.Channel11.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel11.Intensity += i.Channel11.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel12.Name = i.Channel12.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel12.Mz = i.Channel12.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel12.Intensity += i.Channel12.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel13.Name = i.Channel13.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel13.Mz = i.Channel13.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel13.Intensity += i.Channel13.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel14.Name = i.Channel14.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel14.Mz = i.Channel14.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel14.Intensity += i.Channel14.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel15.Name = i.Channel15.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel15.Mz = i.Channel15.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel15.Intensity += i.Channel15.Intensity

					evi.Proteins[j].PhosphoTotalLabels.Channel16.Name = i.Channel16.Name
					evi.Proteins[j].PhosphoTotalLabels.Channel16.Mz = i.Channel16.Mz
					evi.Proteins[j].PhosphoTotalLabels.Channel16.Intensity += i.Channel16.Intensity

					//if k.IsNondegenerateEvidence {
					if k.IsUnique {
						evi.Proteins[j].PhosphoUniqueLabels.Channel1.Name = i.Channel1.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel1.Mz = i.Channel1.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel1.Intensity += i.Channel1.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel2.Name = i.Channel2.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel2.Mz = i.Channel2.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel2.Intensity += i.Channel2.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel3.Name = i.Channel3.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel3.Mz = i.Channel3.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel3.Intensity += i.Channel3.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel4.Name = i.Channel4.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel4.Mz = i.Channel4.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel4.Intensity += i.Channel4.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel5.Name = i.Channel5.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel5.Mz = i.Channel5.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel5.Intensity += i.Channel5.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel6.Name = i.Channel6.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel6.Mz = i.Channel6.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel6.Intensity += i.Channel6.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel7.Name = i.Channel7.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel7.Mz = i.Channel7.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel7.Intensity += i.Channel7.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel8.Name = i.Channel8.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel8.Mz = i.Channel8.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel8.Intensity += i.Channel8.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel9.Name = i.Channel9.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel9.Mz = i.Channel9.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel9.Intensity += i.Channel9.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel10.Name = i.Channel10.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel10.Mz = i.Channel10.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel10.Intensity += i.Channel10.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel11.Name = i.Channel11.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel11.Mz = i.Channel11.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel11.Intensity += i.Channel11.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel12.Name = i.Channel12.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel12.Mz = i.Channel12.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel12.Intensity += i.Channel12.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel13.Name = i.Channel13.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel13.Mz = i.Channel13.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel13.Intensity += i.Channel13.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel14.Name = i.Channel14.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel14.Mz = i.Channel14.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel14.Intensity += i.Channel14.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel15.Name = i.Channel15.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel15.Mz = i.Channel15.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel15.Intensity += i.Channel15.Intensity

						evi.Proteins[j].PhosphoUniqueLabels.Channel16.Name = i.Channel16.Name
						evi.Proteins[j].PhosphoUniqueLabels.Channel16.Mz = i.Channel16.Mz
						evi.Proteins[j].PhosphoUniqueLabels.Channel16.Intensity += i.Channel16.Intensity
					}

					if k.IsURazor {
						evi.Proteins[j].PhosphoURazorLabels.Channel1.Name = i.Channel1.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel1.Mz = i.Channel1.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel1.Intensity += i.Channel1.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel2.Name = i.Channel2.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel2.Mz = i.Channel2.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel2.Intensity += i.Channel2.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel3.Name = i.Channel3.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel3.Mz = i.Channel3.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel3.Intensity += i.Channel3.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel4.Name = i.Channel4.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel4.Mz = i.Channel4.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel4.Intensity += i.Channel4.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel5.Name = i.Channel5.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel5.Mz = i.Channel5.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel5.Intensity += i.Channel5.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel6.Name = i.Channel6.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel6.Mz = i.Channel6.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel6.Intensity += i.Channel6.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel7.Name = i.Channel7.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel7.Mz = i.Channel7.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel7.Intensity += i.Channel7.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel8.Name = i.Channel8.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel8.Mz = i.Channel8.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel8.Intensity += i.Channel8.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel9.Name = i.Channel9.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel9.Mz = i.Channel9.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel9.Intensity += i.Channel9.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel10.Name = i.Channel10.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel10.Mz = i.Channel10.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel10.Intensity += i.Channel10.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel11.Name = i.Channel11.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel11.Mz = i.Channel11.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel11.Intensity += i.Channel11.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel12.Name = i.Channel12.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel12.Mz = i.Channel12.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel12.Intensity += i.Channel12.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel13.Name = i.Channel13.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel13.Mz = i.Channel13.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel13.Intensity += i.Channel13.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel14.Name = i.Channel14.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel14.Mz = i.Channel14.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel14.Intensity += i.Channel14.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel15.Name = i.Channel15.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel15.Mz = i.Channel15.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel15.Intensity += i.Channel15.Intensity

						evi.Proteins[j].PhosphoURazorLabels.Channel16.Name = i.Channel16.Name
						evi.Proteins[j].PhosphoURazorLabels.Channel16.Mz = i.Channel16.Mz
						evi.Proteins[j].PhosphoURazorLabels.Channel16.Intensity += i.Channel16.Intensity
					}
				}

			}
		}
	}

	return evi
}

// NormToTotalProteins calculates the protein level normalization based on total proteins
func NormToTotalProteins(evi rep.Evidence) rep.Evidence {

	var topValue float64
	var channelSum = [16]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var normFactors = [16]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// sum TMT singal for each column
	for _, i := range evi.Proteins {
		channelSum[0] += i.URazorLabels.Channel1.Intensity
		channelSum[1] += i.URazorLabels.Channel2.Intensity
		channelSum[2] += i.URazorLabels.Channel3.Intensity
		channelSum[3] += i.URazorLabels.Channel4.Intensity
		channelSum[4] += i.URazorLabels.Channel5.Intensity
		channelSum[5] += i.URazorLabels.Channel6.Intensity
		channelSum[6] += i.URazorLabels.Channel7.Intensity
		channelSum[7] += i.URazorLabels.Channel8.Intensity
		channelSum[8] += i.URazorLabels.Channel9.Intensity
		channelSum[9] += i.URazorLabels.Channel10.Intensity
		channelSum[10] += i.URazorLabels.Channel11.Intensity
		channelSum[11] += i.URazorLabels.Channel12.Intensity
		channelSum[12] += i.URazorLabels.Channel13.Intensity
		channelSum[13] += i.URazorLabels.Channel14.Intensity
		channelSum[14] += i.URazorLabels.Channel15.Intensity
		channelSum[15] += i.URazorLabels.Channel16.Intensity
	}

	// find the higest value amongst channels
	for _, i := range channelSum {
		if i > topValue {
			topValue = i
		}
	}

	// calculate normalizing factors
	for i := range channelSum {
		normFactors[i] = channelSum[i] / topValue
	}

	// multiply each protein TMT set by the factors to get normalized values
	for _, i := range evi.Proteins {
		i.URazorLabels.Channel1.Intensity *= normFactors[0]
		i.URazorLabels.Channel2.Intensity *= normFactors[1]
		i.URazorLabels.Channel3.Intensity *= normFactors[2]
		i.URazorLabels.Channel4.Intensity *= normFactors[3]
		i.URazorLabels.Channel5.Intensity *= normFactors[4]
		i.URazorLabels.Channel6.Intensity *= normFactors[5]
		i.URazorLabels.Channel7.Intensity *= normFactors[6]
		i.URazorLabels.Channel8.Intensity *= normFactors[7]
		i.URazorLabels.Channel9.Intensity *= normFactors[8]
		i.URazorLabels.Channel10.Intensity *= normFactors[9]
		i.URazorLabels.Channel11.Intensity *= normFactors[10]
		i.URazorLabels.Channel12.Intensity *= normFactors[11]
		i.URazorLabels.Channel13.Intensity *= normFactors[12]
		i.URazorLabels.Channel14.Intensity *= normFactors[13]
		i.URazorLabels.Channel15.Intensity *= normFactors[14]
		i.URazorLabels.Channel16.Intensity *= normFactors[15]
	}

	return evi
}

func calculateRatios(evi rep.Evidence) rep.Evidence {

	var psmSum = make(map[string]float64)
	var psmLog2 = make(map[string]float64)

	// calculate the sum of all intensities for each PSM and then the log2 from the intensities
	for i := range evi.PSM {
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel1.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel2.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel3.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel4.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel5.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel6.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel7.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel8.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel9.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel10.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel11.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel12.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel13.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel14.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel15.Intensity
		psmSum[evi.PSM[i].Spectrum] += evi.PSM[i].Labels.Channel16.Intensity

		psmLog2[evi.PSM[i].Spectrum] = math.Log2(psmSum[evi.PSM[i].Spectrum])
	}

	return evi
}
