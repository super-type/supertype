package main

import (
	"github.com/super-type/supertype/pkg/cmd/supertype"
)

func main() {
	// TODO should we have this be err := supertype.RunApplication and check for errors here?
	// TODO should CheckError just be in the supertype package?
	supertype.RunApplication()
	// cmd.CheckError(err)
}
