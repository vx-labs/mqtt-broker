package cli

var BuiltVersion = "snapshot"

// Version returns the version variable, overriden at build
func Version() string {
	return BuiltVersion
}
