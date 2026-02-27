package shortener

import "hash/fnv"

func hashURL(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func encodeBase62(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}

	var result []byte
	for n > 0 {
		result = append(result, alphabet[n%62])
		n /= 62
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func shortenURL(url string) string {
	hash := hashURL(url)
	short := encodeBase62(hash)

	if len(short) > 8 {
		short = short[:8]
	}

	return short
}
