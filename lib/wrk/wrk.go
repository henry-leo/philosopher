package wrk

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pierrre/archivefile/zip"
	"github.com/nesvilab/philosopher/lib/msg"
	"github.com/nesvilab/philosopher/lib/gth"
	"github.com/nesvilab/philosopher/lib/met"
	"github.com/nesvilab/philosopher/lib/sys"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

// Run is the workspace main entry point
func Run(Version, Build string, b, c, i, n bool) {

	if n == false {
		gth.UpdateChecker(Version, Build)
	}

	if (i == true && b == true && c == true) || (i == true && b == true) || (i == true && c == true) || (c == true && b == true) {
		msg.Custom(errors.New("this command accepts only one parameter"), "fatal")
	}

	if i == true {

		logrus.Info("Creating workspace")
		Init(Version, Build)

	} else if b == true {

		logrus.Info("Creating backup")
		Backup()

	} else if c == true {

		logrus.Info("Removing workspace")
		Clean()
	}

	return
}

// Init creates a new workspace
func Init(version, build string) {

	var m met.Data

	b, _ := ioutil.ReadFile(sys.Meta())

	msgpack.Unmarshal(b, &m)

	if len(m.UUID) > 1 && len(m.Home) > 1 {
		msg.OverwrittingMeta(errors.New(""), "warning")
	} else {

		dir, e := os.Getwd()
		if e != nil {
			msg.GettingLocalDir(e, "warning")
		}

		da := met.New(dir)

		da.Version = version
		da.Build = build

		os.Mkdir(da.MetaDir, sys.FilePermission())
		if _, e := os.Stat(sys.MetaDir()); os.IsNotExist(e) {
			msg.CreatingMetaDirectory(e, "fatal")
		}

		if runtime.GOOS == sys.Windows() {
			HideFile(sys.MetaDir())
		}

		os.Mkdir(da.Temp, sys.FilePermission())
		if _, e := os.Stat(da.Temp); os.IsNotExist(e) {
			msg.LocatingTemDirecotry(e, "fatal")
		}

		da.Home = fmt.Sprintf("%s", da.Home)
		da.MetaDir = fmt.Sprintf("%s", da.MetaDir)

		da.Serialize()
	}

	return
}

// Backup collects all binary files from the workspace and zips them
func Backup() {

	// this is a soft verification just to see if there is any existing file
	var m met.Data
	_, e := ioutil.ReadFile(sys.Meta())
	if e != nil {
		msg.ReadFile(e, "warning")
	}

	if len(m.UUID) < 1 && len(m.Home) < 1 {
		msg.LocatingMetaDirecotry(errors.New(""), "error")
	}

	var name string
	if m.OS == sys.Windows() {
		name = "backup.zip"
	} else {
		t := time.Now()
		timestamp := t.Format(time.RFC3339)
		name = fmt.Sprintf("%s-%s.zip", m.ProjectName, timestamp)
	}
	outFilePath := filepath.Join(m.Home, name)

	progress := func(archivePath string) {
	}

	e = zip.ArchiveFile(sys.MetaDir(), outFilePath, progress)
	if e != nil {
		msg.ArchivingMetaDirecotry(e, "error")
	}

	return
}

// Clean deletes all meta data and the workspace directory
func Clean() {

	// this is a soft verification just to see if there is any existing file
	var d met.Data
	_, e := ioutil.ReadFile(sys.Meta())
	if e != nil {
		msg.ReadFile(e, "warning")
	}

	e = os.RemoveAll(sys.MetaDir())
	if e != nil {
		msg.DeletingMetaDirecotry(e, "warning")
	}

	if len(d.Temp) > 0 {
		e := os.RemoveAll(d.Temp)
		if e != nil {
			msg.DeletingMetaDirecotry(e, "warning")
		}
	}

	return
}
