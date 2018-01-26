package amagi

import (
	"testing"
)

func TestInfo(t *testing.T) {
	Info("testing info...")
}

func TestWarning(t *testing.T) {
	Warn("testing warning...")
}

func TestError(t *testing.T) {
	Error("testing error..")
}

func TestFatal(t *testing.T) {
	Fatal("testing fatal..")
}

func TestAllLogFunc(t *testing.T) {
	Info("testing info...")
	Warn("testing warn...")
	Error("testing error...")
	Fatal("testing fatal...")
}
