package ad_cutter

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestWindows(t *testing.T) {
	isWindows := isWindows()
	t.Logf("current os is windows : %v", isWindows)
}

func TestFFMPEG(t *testing.T) {
	output, err := generateRawData("ADN-499-C.mp4")
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("output file: %s", output)
	}
}

func TestDirPermission(t *testing.T) {
	_, err := dirPermission("./")
	if err != nil {
		t.Error(err.Error())
	}

	_, err = dirPermission("/usr")
	if err == nil {
		t.Errorf("%s should not has write permission", "/usr")
	}
}

func TestAddWritePermission(t *testing.T) {
	err := tryAddWritePermission("./")
	if err != nil {
		t.Error(err.Error())
	}

	err = tryAddWritePermission("/usr")
	if err == nil {
		t.Errorf("/usr should not be added write permission by current user")
	}
}

func TestCMDError(t *testing.T) {
	cmd := exec.Command("ffprobe", "-i", "ADN-499-C.mp4")
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		t.Error(errOut.String())
	}
}
