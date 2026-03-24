package main

import (
	"fmt"

	greet "github.com/jamesstocktonj1/componentize-sdk/examples/greeting/gen/export_jamesstocktonj1_componentize_sdk_examples_greeting_greet"
	_ "github.com/jamesstocktonj1/componentize-sdk/examples/greeting/gen/wit_exports"
)

func init() {
	greet.SetGreet(func(s string) string {
		return fmt.Sprintf("Hello, %s!", s)
	})
}

func main() {}
