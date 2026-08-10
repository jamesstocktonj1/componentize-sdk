package export_wasi_cli_run

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

var run func() error = func() error {
	panic("not implemented")
}

func SetRunFunction(f func() error) {
	run = f
}

func Run() witTypes.Result[witTypes.Unit, witTypes.Unit] {
	if err := run(); err != nil {
		return witTypes.Err[witTypes.Unit, witTypes.Unit](witTypes.Unit{})
	}
	return witTypes.Ok[witTypes.Unit, witTypes.Unit](witTypes.Unit{})
}
