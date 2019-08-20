package met

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/prvst/philosopher/lib/err"

	"github.com/prvst/philosopher/lib/sys"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
)

// Data is the global parameter container
type Data struct {
	UUID           string
	Home           string
	Temp           string
	MetaFile       string
	MetaDir        string
	DB             string
	OS             string
	Arch           string
	Distro         string
	TimeStamp      string
	Version        string
	Build          string
	ProjectName    string
	SearchEngine   string
	Msconvert      Msconvert
	Idconvert      Idconvert
	Database       Database
	MSFragger      MSFragger
	Comet          Comet
	PeptideProphet PeptideProphet
	InterProphet   InterProphet
	ProteinProphet ProteinProphet
	PTMProphet     PTMProphet
	Filter         Filter
	Quantify       Quantify
	Cluster        Cluster
	Abacus         Abacus
	Report         Report
	TMTIntegrator  TMTIntegrator
	Index          Index
	Pipeline       Pipeline
}

// Msconvert options and parameters
type Msconvert struct {
	Output                  string
	Format                  string
	MZBinaryEncoding        string
	IntensityBinaryEncoding string
	NoIndex                 bool
	Zlib                    bool
}

// Idconvert optioons and parameters
type Idconvert struct {
	Format string
}

// Database options and parameters
type Database struct {
	ID     string `yaml:"id"`
	Annot  string `yaml:"protein_database"`
	Enz    string `yaml:"enzyme"`
	Tag    string `yaml:"decoy_tag"`
	Add    string `yaml:"add"`
	Custom string `yaml:"custom"`
	Crap   bool   `yaml:"contam"`
	Rev    bool   `yaml:"reviewed"`
	Iso    bool   `yaml:"isoform"`
	NoD    bool   `yaml:"nodecoys"`
}

// Comet options and parameters
type Comet struct {
	Param        string `yaml:"param"`
	ParamFile    []byte
	RawExtension string `yaml:"raw"`
	RawFiles     []string
	Print        bool
	NoIndex      bool `yaml:"noindex"`
}

// MSFragger options and parameters
type MSFragger struct {
	JarPath                            string  `yaml:"path"`
	Memmory                            int     `yaml:"memmory"`
	Threads                            int     `yaml:"num_threads"`
	RawExtension                       string  `yaml:"raw"`
	DatabaseName                       string  `yaml:"database_name"`
	PrecursorMassLower                 int     `yaml:"precursor_mass_lower"`
	PrecursorMassUpper                 int     `yaml:"precursor_mass_upper"`
	PrecursorMassUnits                 int     `yaml:"precursor_mass_units"`
	PrecursorTrueTolerance             int     `yaml:"precursor_true_tolerance"`
	PrecursorTrueUnits                 int     `yaml:"precursor_true_units"`
	FragmentMassTolerance              float64 `yaml:"fragment_mass_tolerance"`
	FragmentMassUnits                  int     `yaml:"fragment_mass_units"`
	CalibrateMass                      int     `yaml:"calibrate_mass"`
	EvaluateMassCalibration            int     `yaml:"evaluate_mass_calibration"`
	IsotopeError                       string  `yaml:"isotope_error"`
	MassOffsets                        int     `yaml:"mass_offsets"`
	PrecursorMassMode                  string  `yaml:"precursor_mass_mode"`
	LocalizeDeltaMass                  int     `yaml:"localize_delta_mass"`
	DeltaMassExcludeRanges             string  `yaml:"delta_mass_exclude_ranges"`
	FragmentIonSeries                  string  `yaml:"fragment_ion_series"`
	SearchEnzymeName                   string  `yaml:"search_enzyme_name"`
	SearchEnzymeCutafter               string  `yaml:"search_enzyme_cutafter"`
	SearchEnzymeButNotAfter            string  `yaml:"search_enzyme_butnotafter"`
	NumEnzymeTermini                   int     `yaml:"num_enzyme_termini"`
	AllowedMissedCleavage              int     `yaml:"allowed_missed_cleavage"`
	ClipNTermM                         int     `yaml:"clip_nTerm_M"`
	AllowMultipleVariableModsOnResidue int     `yaml:"allow_multiple_variable_mods_on_residue"`
	MaxVariableModsPerMod              int     `yaml:"max_variable_mods_per_mod"`
	MaxVariableModsCombinations        int     `yaml:"max_variable_mods_combinations"`
	OutputFileExtension                string  `yaml:"output_file_extension"`
	OutputFormat                       string  `yaml:"output_format"`
	OutputReportTopN                   int     `yaml:"output_report_topN"`
	OutputMaxExpect                    int     `yaml:"output_max_expect"`
	ReportAlternativeProteins          int     `yaml:"report_alternative_proteins"`
	OverrideCharge                     int     `yaml:"override_charge"`
	PrecursorCharge                    string  `yaml:"precursor_charge"`
	DigestMinLength                    int     `yaml:"digest_min_length"`
	DigestMaxLength                    int     `yaml:"digest_max_length"`
	DigestMassRange                    string  `yaml:"digest_mass_range"`
	MaxFragmentCharge                  int     `yaml:"max_fragment_charge"`
	TrackZeroTopN                      int     `yaml:"track_zero_topN"`
	ZeroBinAcceptExpect                int     `yaml:"zero_bin_accept_expect"`
	ZeroBinMultExpect                  int     `yaml:"zero_bin_mult_expect"`
	AddTopNComplementary               int     `yaml:"add_topN_complementary"`
	MinimumPeaks                       int     `yaml:"minimum_peaks"`
	UseTopNPeaks                       int     `yaml:"use_topN_peaks"`
	MinFragmentsModelling              int     `yaml:"min_fragments_modelling"`
	MinMatchedFragments                int     `yaml:"min_matched_fragments"`
	MinimumRatio                       float64 `yaml:"minimum_ratio"`
	ClearMzRange                       string  `yaml:"clear_mz_range"`
	VariableMod01                      string  `yaml:"variable_mod_01"`
	VariableMod02                      string  `yaml:"variable_mod_02"`
	VariableMod03                      string  `yaml:"variable_mod_03"`
	VariableMod04                      string  `yaml:"variable_mod_04"`
	VariableMod05                      string  `yaml:"variable_mod_05"`
	VariableMod06                      string  `yaml:"variable_mod_06"`
	VariableMod07                      string  `yaml:"variable_mod_07"`
	AddCysteine                        float64 `yaml:"add_C_cysteine"`
	AddCtermPeptide                    float64 `yaml:"add_Cterm_peptide"`
	AddCtermProtein                    float64 `yaml:"add_Cterm_protein"`
	AddAsparticAcid                    float64 `yaml:"add_D_aspartic_acid"`
	AddGlutamicAcid                    float64 `yaml:"add_E_glutamic_acid"`
	AddPhenylAlnine                    float64 `yaml:"add_F_phenylalanine"`
	AddGlycine                         float64 `yaml:"add_G_glycine"`
	AddHistidine                       float64 `yaml:"add_H_histidine"`
	AddIsoleucine                      float64 `yaml:"add_I_isoleucine"`
	AddLysine                          float64 `yaml:"add_K_lysine"`
	AddLeucine                         float64 `yaml:"add_L_leucine"`
	AddMethionine                      float64 `yaml:"add_M_methionine"`
	AddAsparagine                      float64 `yaml:"add_N_asparagine"`
	AddNTermPeptide                    float64 `yaml:"add_Nterm_peptide"`
	AddNtermProteine                   float64 `yaml:"add_Nterm_protein"`
	AddProline                         float64 `yaml:"add_P_proline"`
	AddGlutamine                       float64 `yaml:"add_Q_glutamine"`
	AddArginine                        float64 `yaml:"add_R_arginine"`
	AddSerine                          float64 `yaml:"add_S_serine"`
	AddThreonine                       float64 `yaml:"add_T_threonine"`
	AddValine                          float64 `yaml:"add_V_valine"`
	AddTryptophan                      float64 `yaml:"add_W_tryptophan"`
	AddTyrosine                        float64 `yaml:"add_Y_tyrosine"`
	Param                              string  `yaml:"param"`
	RawFiles                           []string
	ParamFile                          []byte
}

// PeptideProphet options and parameters
type PeptideProphet struct {
	InputFiles    []string
	FileExtension string  `yaml:"extension"`
	Output        string  `yaml:"output"`
	Database      string  `yaml:"database"`
	Rtcat         string  `yaml:"rtcat"`
	Decoy         string  `yaml:"decoy"`
	Enzyme        string  `yaml:"enzyme"`
	Minpiprob     float64 `yaml:"minpiprob"`
	Minrtprob     float64 `yaml:"minrtprob"`
	Minprob       float64 `yaml:"minprob"`
	Masswidth     float64 `yaml:"masswidth"`
	MinPepLen     int     `yaml:"minpeplen"`
	Clevel        int     `yaml:"clevel"`
	Minpintt      int     `yaml:"minpintt"`
	Minrtntt      int     `yaml:"minrtntt"`
	Combine       bool    `yaml:"combine"`
	Exclude       bool    `yaml:"exclude"`
	Leave         bool    `yaml:"leave"`
	Perfectlib    bool    `yaml:"perfectlib"`
	Icat          bool    `yaml:"icat"`
	Noicat        bool    `yaml:"noicat"`
	Zero          bool    `yaml:"zero"`
	Accmass       bool    `yaml:"accmass"`
	Ppm           bool    `yaml:"ppm"`
	Nomass        bool    `yaml:"nomass"`
	Pi            bool    `yaml:"pi"`
	Rt            bool    `yaml:"rt"`
	Glyc          bool    `yaml:"glyc"`
	Phospho       bool    `yaml:"phospho"`
	Maldi         bool    `yaml:"maldi"`
	Instrwarn     bool    `yaml:"instrwarn"`
	Decoyprobs    bool    `yaml:"decoyprobs"`
	Nontt         bool    `yaml:"nontt"`
	Nonmc         bool    `yaml:"nonmc"`
	Expectscore   bool    `yaml:"expectscore"`
	Nonparam      bool    `yaml:"nonparam"`
	Neggamma      bool    `yaml:"neggamma"`
	Forcedistr    bool    `yaml:"forcedistr"`
	Optimizefval  bool    `yaml:"optimizefval"`
}

// InterProphet options and parameters
type InterProphet struct {
	InputFiles []string
	Output     string  `yaml:"output"`
	Decoy      string  `yaml:"decoy"`
	Cat        string  `yaml:"cat"`
	Threads    int     `yaml:"threads"`
	MinProb    float64 `yaml:"minprob"`
	Length     bool    `yaml:"length"`
	Nofpkm     bool    `yaml:"nofpkm"`
	Nonss      bool    `yaml:"nonss"`
	Nonse      bool    `yaml:"nonse"`
	Nonrs      bool    `yaml:"nonrs"`
	Nonsm      bool    `yaml:"nonsm"`
	Nonsp      bool    `yaml:"nonsp"`
	Sharpnse   bool    `yaml:"sharpnse"`
	Nonsi      bool    `yaml:"nonsi"`
}

// ProteinProphet options and parameters
type ProteinProphet struct {
	InputFiles  []string
	Output      string  `yaml:"output"`
	Minindep    int     `yaml:"minidep"`
	Mufactor    int     `yaml:"mufactor"`
	Maxppmdiff  int     `yaml:"maxppmdiff"`
	Minprob     float64 `yaml:"minprob"`
	ExcludeZ    bool    `yaml:"excludez"`
	Noplot      bool    `yaml:"noplot"`
	Nooccam     bool    `yaml:"noocam"`
	Softoccam   bool    `yaml:"softocam"`
	Icat        bool    `yaml:"icat"`
	Glyc        bool    `yaml:"glyc"`
	Nogroupwts  bool    `yaml:"nogroupwts"`
	NonSP       bool    `yaml:"nonsp"`
	Accuracy    bool    `yaml:"accuracy"`
	Asap        bool    `yaml:"asap"`
	Refresh     bool    `yaml:"refresh"`
	Normprotlen bool    `yaml:"normprotlen"`
	Logprobs    bool    `yaml:"logprobs"`
	Confem      bool    `yaml:"confem"`
	Allpeps     bool    `yaml:"allpeps"`
	Unmapped    bool    `yaml:"unmapped"`
	Noprotlen   bool    `yaml:"noprotlen"`
	Instances   bool    `yaml:"instances"`
	Fpkm        bool    `yaml:"fpkm"`
	Protmw      bool    `yaml:"protmw"`
	Iprophet    bool    `yaml:"iprophet"`
	Asapprophet bool    `yaml:"asapprophet"`
	Delude      bool    `yaml:"delude"`
	Excludemods bool    `yaml:"excludemods"`
}

// PTMProphet options and parameters
type PTMProphet struct {
	InputFiles   []string
	Output       string  `yaml:"output"`
	Mods         string  `yaml:"mods"`
	NIons        string  `yaml:"nions"`
	CIons        string  `yaml:"cions"`
	EM           int     `yaml:"em"`
	FragPPMTol   int     `yaml:"fragppmtol"`
	MaxThreads   int     `yaml:"maxthreads"`
	MaxFragZ     int     `yaml:"maxfragz"`
	Mino         int     `yaml:"mino"`
	MassOffset   int     `yaml:"massoffset"`
	PPMTol       float64 `yaml:"ppmtol"`
	MinProb      float64 `yaml:"minprob"`
	Static       bool    `yaml:"static"`
	NoUpdate     bool    `yaml:"noupdate"`
	KeepOld      bool    `yaml:"keepold"`
	Verbose      bool    `yaml:"verbose"`
	MassDiffMode bool    `yaml:"massdiffmode"`
	Lability     bool    `yaml:"lability"`
	Direct       bool    `yaml:"direct"`
	Ifrags       bool    `yaml:"ifrags"`
	Autodirect   bool    `yaml:"autodirect"`
	NoMinoFactor bool    `yaml:"nominofactor"`
}

// Filter options and parameters
type Filter struct {
	Pex      string  `yaml:"pepxml"`
	Pox      string  `yaml:"protxml"`
	Tag      string  `yaml:"tag"`
	PsmFDR   float64 `yaml:"psmFDR"`
	PepFDR   float64 `yaml:"peptideFDR"`
	IonFDR   float64 `yaml:"ionFDR"`
	PtFDR    float64 `yaml:"proteinFDR"`
	ProtProb float64 `yaml:"proteinProbability"`
	PepProb  float64 `yaml:"peptideProbability"`
	Weight   float64 `yaml:"peptideWeight"`
	Model    bool    `yaml:"models"`
	Razor    bool    `yaml:"razor"`
	Picked   bool    `yaml:"picked"`
	Seq      bool    `yaml:"sequential"`
	Cap      bool    `yaml:"cappedsequential"`
	Mapmods  bool    `yaml:"mapMods"`
	Fo       bool
}

// Quantify options and parameters
type Quantify struct {
	Format     string  `yaml:"format"`
	Dir        string  `yaml:"dir"`
	Brand      string  `yaml:"brand"`
	Plex       string  `yaml:"plex"`
	ChanNorm   string  `yaml:"chanNorm"`
	Annot      string  `yaml:"annotation"`
	Level      int     `yaml:"level"`
	RTWin      float64 `yaml:"retentionTimeWindow"`
	PTWin      float64 `yaml:"peakTimeWindow"`
	Tol        float64 `yaml:"tolerance"`
	Purity     float64 `yaml:"purity"`
	MinProb    float64 `yaml:"minprob"`
	RemoveLow  float64 `yaml:"removeLow"`
	Isolated   bool    `yaml:"isolated"`
	IntNorm    bool    `yaml:"intNorm"`
	Unique     bool    `yaml:"uniqueOnly"`
	BestPSM    bool    `yaml:"bestPSM"`
	LabelNames map[string]string
}

// Abacus options ad parameters
type Abacus struct {
	Tag      string  `yaml:"tag"`
	ProtProb float64 `yaml:"proteinProbability"`
	PepProb  float64 `yaml:"peptideProbability"`
	Peptide  bool    `yaml:"peptide"`
	Protein  bool    `yaml:"protein"`
	Razor    bool    `yaml:"razor"`
	Picked   bool    `yaml:"picked"`
	Labels   bool    `yaml:"labels"`
	Unique   bool    `yaml:"uniqueOnly"`
	Reprint  bool    `yaml:"reprint"`
}

// Cluster options and parameters
type Cluster struct {
	UID   string  `yaml:"organismUniProtID"`
	Level float64 `yaml:"level"`
}

// Report options and parameters
type Report struct {
	Decoys  bool `yaml:"withDecoys"`
	MSstats bool `yaml:"msstats"`
	MZID    bool `yaml:"mzID"`
}

// TMTIntegrator options and parameters
type TMTIntegrator struct {
	JarPath   string `yaml:"path"`
	Memmory   int    `yaml:"memmory"`
	Param     string `yaml:"param"`
	Files     []string
	ParamFile []byte
}

// Index options and parameters
type Index struct {
	Spectra string
}

// Pipeline options and parameters
type Pipeline struct {
	Directives string
	Print      bool
}

// New initializes the structure with the system information needed
// to run all the follwing commands
func New(h string) Data {

	var d Data

	var fmtuuid, _ = uuid.NewV4()
	var uuid = fmt.Sprintf("%s", fmtuuid)
	d.UUID = uuid

	d.OS = runtime.GOOS
	d.Arch = runtime.GOARCH

	distro := sys.GetLinuxFlavor()

	d.Distro = distro

	d.Home = h
	d.ProjectName = string(filepath.Base(h))

	d.MetaFile = d.Home + string(filepath.Separator) + sys.Meta()
	d.MetaDir = d.Home + string(filepath.Separator) + sys.MetaDir()

	d.DB = d.Home + string(filepath.Separator) + sys.DBBin()

	temp := sys.GetTemp()
	temp += string(filepath.Separator) + uuid
	d.Temp = temp

	t := time.Now()
	d.TimeStamp = t.Format(time.RFC3339)

	return d
}

// CleanTemp removes all files from the given temp directory
func CleanTemp(dir string) {

	files, e := filepath.Glob(filepath.Join(dir, "*"))
	if e != nil {
		err.Custom(e, "fatal")
	}
	for _, file := range files {
		e = os.RemoveAll(file)
		if e != nil {
			err.Custom(e, "fatal")
		}
	}

	return
}

// Serialize converts the whole structure to a gob file
func (d *Data) Serialize() {

	b, e := msgpack.Marshal(&d)
	if e != nil {
		err.MarshalFile(e, "fatal")
	}

	e = ioutil.WriteFile(sys.Meta(), b, sys.FilePermission())
	if e != nil {
		err.WriteFile(e, "fatal")
	}

	return
}

// Restore reads philosopher results files and restore the data sctructure
func (d *Data) Restore(f string) {

	b, e1 := ioutil.ReadFile(f)

	e2 := msgpack.Unmarshal(b, &d)

	if e1 != nil && e2 != nil && len(d.UUID) < 1 {
		err.Custom(errors.New("The current directory has no Workspace"), "warning")
	}

	// checks if the temp is still there, if not recreate it
	if _, err := os.Stat(d.Temp); os.IsNotExist(err) {
		os.Mkdir(d.Temp, sys.FilePermission())
	}

	return
}

// FunctionInitCheckUp does initilization checkup and verification if meta and temp folders are up.
// In case not, meta trows an error and folder is created.
func (d Data) FunctionInitCheckUp() {

	if len(d.UUID) < 1 && len(d.Home) < 1 {
		err.WorkspaceNotFound(errors.New(""), "warning")
	}

	if _, e := os.Stat(d.Temp); os.IsNotExist(e) && len(d.UUID) > 0 {
		os.Mkdir(d.Temp, sys.FilePermission())
		err.LocatingTemDirecotry(e, "warning")
	}

	return
}
