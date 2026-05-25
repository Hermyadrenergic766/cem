package main

// suggestTool — KnownTools içinde input'a en yakın anahtarı (Levenshtein
// mesafesi ≤ 2) döndürür. Tam eşleşme önceliklidir; yoksa ortografik
// olarak benzer tool önerilir. Bulunamazsa "" döner.
//
// Örnek: "cluade" → "claude", "gpt5" → "gpt", "cursr" → "cursor".
func suggestTool(input string) string {
	if input == "" {
		return ""
	}
	if _, ok := KnownTools[input]; ok {
		return input
	}
	best := ""
	bestDist := 3 // 3 ve üstü çok uzak — eşik
	for key := range KnownTools {
		d := levenshtein(input, key)
		if d < bestDist {
			bestDist = d
			best = key
		}
	}
	return best
}

// levenshtein — standart edit distance. m,n küçük olduğu için (4 tool, kısa
// isimler) DP O(m*n) yeterli.
func levenshtein(a, b string) int {
	m, n := len(a), len(b)
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = minInt(
				prev[j]+1,        // deletion
				curr[j-1]+1,      // insertion
				prev[j-1]+cost,   // substitution
			)
		}
		prev, curr = curr, prev
	}
	return prev[n]
}

func minInt(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
