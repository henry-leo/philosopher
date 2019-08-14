package cdhit

import (
	"errors"
	"io/ioutil"

	"github.com/prvst/philosopher/lib/err"

	"github.com/prvst/philosopher/lib/sys"
)

// Unix64 ...
func Unix64(unix64 string) {

	bin, e := Asset("cd-hit")
	e = ioutil.WriteFile(unix64, bin, sys.FilePermission())

	if e != nil {
		err.DeployAsset(errors.New("CD-HIT"))
	}

	return
}
