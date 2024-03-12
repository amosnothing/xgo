package test

import (
	"testing"
)

// go test -run TestTrapLocal -v ./test
func TestTrapLocal(t *testing.T) {
	origExpect := "hello\nhello\n"
	expectOut := "global trap: main\nglobal trap: A\nlocal trap from A: printHello\nglobal trap: printHello\nhello\nglobal trap: B\nlocal trap from B: printHello\nglobal trap: printHello\nhello\n"
	testTrap(t, "./testdata/trap_local", origExpect, expectOut)
}
