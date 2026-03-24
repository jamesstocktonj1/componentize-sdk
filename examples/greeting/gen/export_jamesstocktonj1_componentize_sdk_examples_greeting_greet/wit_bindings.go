package export_jamesstocktonj1_componentize_sdk_examples_greeting_greet

var greet func(string) string

func Greet(name string) string {
	return greet(name)
}

func SetGreet(f func(string) string) {
	greet = f
}
