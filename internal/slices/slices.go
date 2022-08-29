package slices

// Contains reports whether v is within s.
func Contains[T comparable](s []T, v T) bool {
	return Index(s, v) >= 0
}

// Index returns the index of the first instance of v in s, or -1 if v is not
// present in s.
func Index[T comparable](s []T, v T) int {
	for i := range s {
		if s[i] == v {
			return i
		}
	}
	return -1
}

// Equal returns whether a and b's contents are identical.
func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Unique filters out repeated elements in a slice. If s is sorted, Unique
// returns a slice with no duplicate entries.
func Unique[T comparable](s []T) []T {
	if len(s) == 0 {
		return nil
	}

	previous := s[0]
	t := []T{previous}

	for i := 1; i < len(s); i++ {
		if s[i] != previous {
			t = append(t, s[i])
		}
		previous = s[i]
	}

	return t
}

func Duplicates[T comparable](s []T) int {
	uniq := make(map[T]struct{})

	for _, v := range s {
		uniq[v] = struct{}{}
	}

	return len(s) - len(uniq)
}
