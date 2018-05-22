package rotate

import "strings"

// Rotate string by a sep
func Rotate(raw, sep string) string {
	items := strings.Split(raw, sep)

	for i := 0; i < (len(items)+1)/2; i++ {
		items[i], items[len(items)-1-i] = items[len(items)-1-i], items[i]
	}
	return strings.Join(items, sep)
}
