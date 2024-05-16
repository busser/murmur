package extern

import (
	"github.com/busser/murmur/internal/murmur"
)

func ResolveAll(vars map[string]string) (map[string]string, error) {
	return murmur.ResolveAll(vars)
}
