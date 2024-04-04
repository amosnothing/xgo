package patch

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"
	xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"
	"cmd/internal/src"
	"strings"
)

const xgoRuntimePkgPrefix = xgo_ctxt.XgoRuntimePkg + "/"
const xgoTestPkgPrefix = xgo_ctxt.XgoModule + "/test/"
const xgoRuntimeTrapPkg = xgoRuntimePkgPrefix + "trap"

// accepts interface{} as argument
const xgoOnTestStart = "__xgo_on_test_start"

const XgoLinkSetTrap = "__xgo_link_set_trap"
const XgoTrapForGenerated = "__xgo_trap_for_generated"
const setTrap = "__xgo_set_trap"

// only allowed from reflect
const reflectSetImpl = "__xgo_set_all_method_by_name_impl"

var linkMap = map[string]string{
	"__xgo_link_getcurg":                      "__xgo_getcurg",
	"__xgo_link_set_trap":                     setTrap,
	xgo_syntax.XgoLinkTrapForGenerated:        XgoTrapForGenerated,
	"__xgo_link_init_finished":                "__xgo_init_finished",
	"__xgo_link_on_init_finished":             "__xgo_on_init_finished",
	"__xgo_link_on_gonewproc":                 "__xgo_on_gonewproc",
	"__xgo_link_on_goexit":                    "__xgo_on_goexit",
	"__xgo_link_on_test_start":                xgoOnTestStart,
	"__xgo_link_get_test_starts":              "__xgo_get_test_starts",
	"__xgo_link_retrieve_all_funcs_and_clear": "__xgo_retrieve_all_funcs_and_clear",
	"__xgo_link_peek_panic":                   "__xgo_peek_panic",
	"__xgo_link_mem_equal":                    "__xgo_mem_equal",
	"__xgo_link_get_pc_name":                  "__xgo_get_pc_name",
	xgo_syntax.XgoLinkGeneratedRegisterFunc:   "__xgo_register_func",

	// reflect (not enabled)
	// "__xgo_link_set_all_method_by_name_impl": reflectSetImpl,
	// "__xgo_link_get_all_method_by_name":      "__xgo_get_all_method_by_name",
}

// a switch to control link
const disableXgoLink bool = false

func isLinkValid(fnName string, targetName string, pkgPath string) bool {
	if disableXgoLink {
		return false
	}
	safeGenerated := (fnName == xgo_syntax.XgoLinkGeneratedRegisterFunc || fnName == xgo_syntax.XgoLinkTrapForGenerated)
	if safeGenerated {
		// generated by xgo on the fly for every instrumented package
		return true
	}
	isReflect := targetName == reflectSetImpl
	if isReflect {
		// the special reflect
		return pkgPath == "reflect"
	}

	isLinkTrap := fnName == XgoLinkSetTrap
	if isLinkTrap {
		// the special trap
		return pkgPath == xgoRuntimeTrapPkg || strings.HasPrefix(pkgPath, xgoTestPkgPrefix)
	}

	// no special link, must be inside xgoRuntime, or test
	if pkgPath == "testing" {
		return true
	}
	if strings.HasPrefix(pkgPath, xgoRuntimePkgPrefix) {
		return true
	}
	if strings.HasPrefix(pkgPath, xgoTestPkgPrefix) {
		return true
	}
	return false
}

// for go1.20 and above, needs to convert
const needConvertArg = goMajor > 1 || (goMajor == 1 && goMinor >= 20)

func replaceWithRuntimeCall(fn *ir.Func, name string) {
	if false {
		debugReplaceBody(fn)
		// newBody = []ir.Node{debugPrint("replaced body")}
		return
	}
	isRuntime := true
	var runtimeFunc *ir.Name
	if isRuntime {
		runtimeFunc = typecheck.LookupRuntime(name)
	} else {
		// NOTE: cannot reference testing package
		// only runtime is available
		// lookup testing
		testingPkg := findTestingPkg()
		sym := testingPkg.Lookup(name)
		if sym.Def != nil {
			runtimeFunc = sym.Def.(*ir.Name)
		} else {
			runtimeFunc = NewNameAt(fn.Pos(), sym, fn.Type())
			runtimeFunc.Class = ir.PEXTERN
		}
	}
	params := fn.Type().Params()
	results := fn.Type().Results()

	paramNames := getTypeNames(params)
	if name == XgoTrapForGenerated {
		// fill PC by
		getCallerPC := typecheck.LookupRuntime("getcallerpc")
		paramNames[1] = ir.NewCallExpr(fn.Pos(), ir.OCALL, getCallerPC, nil)
	}
	resNames := getTypeNames(results)
	fnPos := fn.Pos()

	if needConvertArg && name == xgoOnTestStart {
		for i, p := range paramNames {
			paramNames[i] = convToEFace(fnPos, p, p.(*ir.Name).Type(), false)
		}
	}

	var callNode ir.Node
	callNode = ir.NewCallExpr(fnPos, ir.OCALL, runtimeFunc, paramNames)
	if len(resNames) > 0 {
		// if len(resNames) == 1 {
		// 	callNode = ir.NewAssignListStmt(fnPos, ir.OAS, resNames, []ir.Node{callNode})
		// } else {
		callNode = ir.NewReturnStmt(fnPos, []ir.Node{callNode})
		// callNode = ir.NewAssignListStmt(fnPos, ir.OAS2, resNames, []ir.Node{callNode})

		// callNode = ir.NewAssignListStmt(fnPos, ir.OAS2, resNames, []ir.Node{callNode})
		// }
	}
	replaceFuncBodyWithPos(base.AutogeneratedPos, fn, []ir.Node{
		// debugPrint("debug getg"),
		callNode,
	})
}

func replaceFuncBody(fn *ir.Func, nodes []ir.Node) {
	replaceFuncBodyWithPos(fn.Pos(), fn, nodes)
}

func replaceFuncBodyWithPos(pos src.XPos, fn *ir.Func, nodes []ir.Node) {
	node := ifConstant(pos, true, nodes, fn.Body)

	fn.Body = []ir.Node{node}
	xgo_record.SetRewrittenBody(fn, fn.Body)
	typeCheckBody(fn)
}
