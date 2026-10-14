package param

import "strings"

// ParseSchemeURI splits a value shaped like "<scheme>://<rest>" into its
// scheme ("cloud", "secret", ...) and reports ok=true. Anything without a
// "://" separator is a plain local path (or any other ordinary string) and
// returns ok=false, unchanged — this is what keeps the file-variable
// lifecycle strictly additive: a value with no "://" is never touched.
func ParseSchemeURI(value string) (scheme string, ok bool) {
	idx := strings.Index(value, "://")
	if idx <= 0 {
		return "", false
	}
	return value[:idx], true
}
