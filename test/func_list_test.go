package test

import (
	"os"
	"os/exec"
	"testing"
)

// go test -run TestFuncList -v ./test
func TestFuncList(t *testing.T) {
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpFile)

	// func_list depends on xgo/runtime, but xgo/runtime is
	// a separate module, so we need to merge them
	// together first
	tmpDir, funcListDir, err := tmpMergeRuntimeAndTest("./testdata/func_list")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = xgoBuild([]string{
		"-o", tmpFile,
		"--project-dir", funcListDir,
		".",
	}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	out, err := exec.Command(tmpFile).Output()
	if err != nil {
		t.Fatalf("%v", err)
	}
	output := string(out)
	// t.Logf("%s", output)

	expectLines := []string{
		"func:strconv FormatBool",
		"func:time Now",
		"func:os MkdirAll",
		"func:fmt Printf",
		"func:strings Split",
		"func:main example",
		"func:main someInt.value",
		"func:main (*someInt).inc",
	}
	expectSequence(t, output, expectLines)
}