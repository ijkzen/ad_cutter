package ad_cutter

import (
	"os"
	"testing"
)

func TestCut(t *testing.T) {
	os.Remove("output.raw")
	result := Cut("ADN-499-C.mp4")
	if result.ErrorMessage != "" {
		t.Error(result.ErrorMessage)
	}
}
