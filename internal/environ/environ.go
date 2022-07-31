package environ

import "strings"

// ToMap takes strings in the form "key=value", as output by os.Environ, and
// returns a corresponding map. It panics if any of the strings are in the wrong
// format.
func ToMap(env []string) map[string]string {
	m := make(map[string]string)

	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) < 2 {
			panic("strings not in the form \"key=value\"")
		}
		k, v := pair[0], pair[1]
		m[k] = v
	}

	return m
}

// ToSlice returns a list of strings in the form "key=value", just like
// os.Environ.
func ToSlice(env map[string]string) []string {
	var s []string

	for k, v := range env {
		s = append(s, k+"="+v)
	}

	return s
}
