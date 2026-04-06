package export_wasi_cli_0_2_6_run

import witTypes "go.bytecodealliance.org/pkg/wit/types"

var function func()

func Run() witTypes.Result[struct{}, struct{}] {
	function()
	return witTypes.Ok[struct{}, struct{}](struct{}{})
}

func SetRunner(f func()) {
	function = f
}
