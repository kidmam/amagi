package helpers

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// CurrentOSVer get current runtime OS
func CurrentOSVer() string {
	return runtime.GOOS
}

// LookPath get bin path to where is available
func LookPath(binName string) (string, error) {
	path, err := exec.LookPath(binName)
	if err != nil {
		log.Fatal(fmt.Sprintf("not found %v", binName))
		return path, err

	}

	return path, nil
}
