package utils

var RegisteredHooks []func()

func RegisterPostInitHook(f func()) {
	RegisteredHooks = append(RegisteredHooks, f)
}
