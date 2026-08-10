package cli

import (
	"github.com/jamesstocktonj1/componentize-sdk/p3/exports/cli/gen/export_wasi_cli_run"
	_ "github.com/jamesstocktonj1/componentize-sdk/p3/exports/cli/gen/wit_exports"
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
