// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code
//
// usage:
//  go run -tags dev ./cmd/xgo test --project-dir runtime/test/debug
//  go run -tags dev ./cmd/xgo test --debug-compile --project-dir runtime/test/debug

package debug

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

type GenericSt[T any] struct {
	Data T
}

func (g GenericSt[T]) GetData(param T) T {
	return param
}

type Inner struct {
}

func TestGeneric(t *testing.T) {
	v := GenericSt[Inner]{}

	var mocked bool
	mock.Patch(v.GetData, func(Inner) Inner {
		mocked = true
		return Inner{}
	})
	v.GetData(Inner{})
	if !mocked {
		t.Fatalf("expected mocked, actually not")
	}
}
