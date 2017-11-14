package data

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/prvst/cmsl/bio"
	"github.com/prvst/cmsl/data/fas"
	"github.com/prvst/cmsl/db"
	"github.com/prvst/cmsl/err"
	"github.com/prvst/philosopher/lib/sys"
	"github.com/vmihailenco/msgpack"
)

// Base main structure
type Base struct {
	UniProtDB string
	CrapDB    string
	TaDeDB    map[string]string
	Records   []db.Record
}

// ProcessDB ...
func (d *Base) ProcessDB(file, decoyTag string) *err.Error {

	fastaMap, e := fas.ParseFile(file)
	if e != nil {
		return e
	}

	for k, v := range fastaMap {

		class, e := db.Classify(k, decoyTag)
		if e != nil {
			return e
		}

		if class == "uniprot" {

			db, e := db.ProcessUniProtKB(k, v, decoyTag)
			if e != nil {
				return e
			}
			d.Records = append(d.Records, db)

		} else if class == "ncbi" {

			db, e := db.ProcessNCBI(k, v, decoyTag)
			if e != nil {
				return e
			}
			d.Records = append(d.Records, db)

		} else if class == "generic" {

			db, e := db.ProcessGeneric(k, v, decoyTag)
			if e != nil {
				return e
			}
			d.Records = append(d.Records, db)

		}

	}

	return nil
}

// Fetch downloads a database file from UniProt
func (d *Base) Fetch(id, temp string, iso, rev bool) *err.Error {

	var query string
	d.UniProtDB = fmt.Sprintf("%s%s%s.fas", temp, string(filepath.Separator), id)

	if rev == true {
		query = fmt.Sprintf("%s%s%s", "http://www.uniprot.org/uniprot/?query=reviewed:yes+AND+proteome:", id, "&format=fasta")
	} else {
		query = fmt.Sprintf("%s%s%s", "http://www.uniprot.org/uniprot/?query=proteome:", id, "&format=fasta")
	}

	if iso == true {
		query = fmt.Sprintf("%s&include=yes", query)
	} else {
		query = fmt.Sprintf("%s&include=no", query)
	}

	// tries to create an output file
	output, e := os.Create(d.UniProtDB)
	if e != nil {
		return &err.Error{Type: err.CannotCreateOutputFile, Class: err.FATA}
	}
	defer output.Close()

	// Tries to query data from Uniprot
	response, e := http.Get(query)
	if e != nil {
		return &err.Error{Type: err.CannotFindUniProtAnnotation, Class: err.FATA, Argument: e.Error()}
	}
	defer response.Body.Close()

	// Tries to download data from Uniprot
	_, e = io.Copy(output, response.Body)
	if e != nil {
		return &err.Error{Type: err.CannotFindUniProtAnnotation, Class: err.FATA, Argument: e.Error()}
	}

	return nil
}

// Create processes the given fasta file and add decoy sequences
func (d *Base) Create(temp, add, enz, tag string, crap bool) *err.Error {

	d.TaDeDB = make(map[string]string)

	dbfile, _ := filepath.Abs(d.UniProtDB)
	db, e := fas.ParseFile(dbfile)
	if e != nil {
		return e
	}

	if len(add) > 0 {
		add, adderr := fas.ParseFile(add)
		if adderr != nil {
			return adderr
		}

		for k, v := range add {
			db[k] = v
		}
	}

	// adding contaminants to database before reversion
	// repeated entries are removed and substituted by contaminants
	if crap == true {

		d.Deploy(temp)

		crapMap, e := fas.ParseFile(d.CrapDB)
		if e != nil {
			return &err.Error{Type: err.CannotParseFastaFile, Class: err.FATA}
		}

		for k, v := range crapMap {
			split := strings.Split(k, "|")
			for i := range db {
				if strings.Contains(i, split[1]) {
					delete(db, i)
				}
			}
			db[k] = v
		}

	}

	var en bio.Enzyme
	en.Synth(enz)
	reg := regexp.MustCompile(en.Pattern)

	for h, s := range db {

		th := ">" + h
		d.TaDeDB[th] = s

		var revPeptides []string
		split := reg.Split(s, -1)
		if len(split) > 1 {
			for i := range split {
				r := reverseSeq(split[i])
				revPeptides = append(revPeptides, r)
			}
		} else {
			r := reverseSeq(s)
			revPeptides = append(revPeptides, r)
		}

		rev := strings.Join(revPeptides, en.Join)
		dh := ">" + tag + h
		d.TaDeDB[dh] = rev
	}

	return nil
}

// Deploy crap file to session folder
func (d *Base) Deploy(temp string) *err.Error {

	d.CrapDB = fmt.Sprintf("%s%scrap.fas", temp, string(filepath.Separator))

	param, e := Asset("crap.fas")
	e = ioutil.WriteFile(d.CrapDB, param, 0644)

	if e != nil {
		return &err.Error{Type: err.CannotDeployCrapDB, Class: err.FATA, Argument: e.Error()}
	}

	return nil
}

// Save fasta file to disk
func (d *Base) Save(home, temp, tag string) *err.Error {

	base := filepath.Base(d.UniProtDB)

	t := time.Now()
	stamp := fmt.Sprintf(t.Format("2006-01-02"))

	workfile := fmt.Sprintf("%s%s%s-td-%s", temp, string(filepath.Separator), stamp, base)
	outfile := fmt.Sprintf("%s%s%s-td-%s", home, string(filepath.Separator), stamp, base)

	// create decoy db file
	file, e := os.Create(workfile)
	if e != nil {
		return &err.Error{Type: err.CannotOpenFile, Class: err.FATA, Argument: "there was an error trying to create the decoy database"}
	}
	defer file.Close()

	var headers []string
	for k := range d.TaDeDB {
		headers = append(headers, k)
	}

	sort.Strings(headers)

	for _, i := range headers {
		line := i + "\n" + d.TaDeDB[i] + "\n"
		_, e = io.WriteString(file, line)
		if e != nil {
			return &err.Error{Type: err.CannotOpenFile, Class: err.FATA, Argument: "there was an error trying to create the database file"}
		}
	}

	sys.CopyFile(workfile, outfile)

	d.ProcessDB(outfile, tag)

	err := d.Serialize()
	if err != nil {
		return err
	}

	return nil
}

// Serialize saves to disk a msgpack verison of the database data structure
func (d *Base) Serialize() *err.Error {

	// create a file
	dataFile, e := os.Create(sys.DBBin())
	if e != nil {
		return &err.Error{Type: err.CannotOpenFile, Class: err.FATA, Argument: "database structure"}
	}

	dataEncoder := msgpack.NewEncoder(dataFile)
	e = dataEncoder.Encode(d)
	if e != nil {
		return &err.Error{Type: err.CannotOpenFile, Class: err.FATA, Argument: e.Error()}
	}
	dataFile.Close()

	return nil
}

// Restore reads philosopher results files and restore the data sctructure
func (d *Base) Restore() *err.Error {

	file, _ := os.Open(sys.DBBin())

	dec := msgpack.NewDecoder(file)
	e := dec.Decode(&d)
	if e != nil {
		return &err.Error{Type: err.CannotOpenFile, Class: err.FATA, Argument: ": database data may be corrupted"}
	}

	return nil
}

// RestoreWithPath reads philosopher results files and restore the data sctructure
func (d *Base) RestoreWithPath(p string) *err.Error {

	var path string

	if strings.Contains(p, string(filepath.Separator)) {
		path = fmt.Sprintf("%s%s", p, sys.DBBin())
	} else {
		path = fmt.Sprintf("%s%s%s", p, string(filepath.Separator), sys.DBBin())
	}

	file, _ := os.Open(path)

	dec := msgpack.NewDecoder(file)
	e := dec.Decode(&d)
	if e != nil {
		return &err.Error{Type: err.CannotOpenFile, Class: err.FATA, Argument: ": database data may be corrupted"}
	}

	return nil
}

// reverseSeq returns its argument string reversed rune-wise left to right.
func reverseSeq(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
