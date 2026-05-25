package main

import (
	"strconv"
	"strings"
)

// semverLess — basit "vX.Y.Z" karşılaştırma. true → a < b.
// "v" öneki opsiyonel. Sayı olmayan parçalar 0 sayılır (pre-release etiketleri
// dikkate alınmaz; cem release'leri saf semver olduğu için yeterli).
func semverLess(a, b string) bool {
	aParts := parseSemver(a)
	bParts := parseSemver(b)
	n := len(aParts)
	if len(bParts) > n {
		n = len(bParts)
	}
	for i := 0; i < n; i++ {
		ai, bi := 0, 0
		if i < len(aParts) {
			ai = aParts[i]
		}
		if i < len(bParts) {
			bi = bParts[i]
		}
		if ai != bi {
			return ai < bi
		}
	}
	return false
}

func parseSemver(s string) []int {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	parts := strings.Split(s, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, _ := strconv.Atoi(p)
		out = append(out, n)
	}
	return out
}
