// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/mock"
)

const A = 20 * time.Second

func TestHello(t *testing.T) {
	mock.PatchByName("github.com/xhd2015/xgo/runtime/test/debug", "A", func() time.Duration {
		return 10 * time.Second
	})
	a := A
	if a != 20*time.Second {
		t.Fatalf("expect patch A failed because current xgo does not resolve operation type, actual: a=%v, want: %v", a, 20*time.Second)
	}
}
