package cli

import (
	"github.com/jamesstocktonj1/componentize-sdk/gen/export_wasi_cli_run"
	_ "github.com/jamesstocktonj1/componentize-sdk/gen/wit_exports"
)

func SetRun(f func()) {
	export_wasi_cli_run.SetRunFunction(func() error {
		f()
		return nil
	})
}

func SetRunE(f func() error) {
	export_wasi_cli_run.SetRunFunction(f)
}
