//go:build ignore

package main

import (
	"fmt"

	"github.com/qatoolist/wowapi/kernel/appmodel"
	"github.com/qatoolist/wowapi/kernel/port"
)

type DummyService interface {
	Foo()
}

func main() {
	// Attempt 1: Direct structural fabrication of a Registrar for another module.
	// This MUST fail compilation because 'owner' and 'compiler' fields are unexported.
	forgedReg := appmodel.Registrar[any]{
		owner: "victim_module",
	}

	// Attempt 2: Direct fabrication with zero-value construction and setting unexported field via reflect
	// (Go compiler blocks direct unexported field assignments at compile-time).
	fmt.Println(forgedReg.Owner())

	// Attempt 3: Try to call the unexported seal() method of another owner's registrar.
	// This MUST fail compilation because 'seal' is unexported.
	forgedReg.seal()

	// Attempt 4: Use port Key with mismatched generic type
	keyInt := port.NewKey[int]("test_port")
	
	// We try to register with a forged registrar
	_ = port.Define(forgedReg, keyInt)
}
