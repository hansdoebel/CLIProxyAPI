//go:build notiktoken

package helps

// approxTokenCounter estimates one token per four bytes. Lean Raspberry Pi
// builds use it so the multi-megabyte tiktoken tables are not linked.
type approxTokenCounter struct{}

func (approxTokenCounter) Count(text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	count := (len(text) + 3) / 4
	if count < 1 {
		count = 1
	}
	return count, nil
}

// LoadTokenCounter ignores kind and returns the approximation counter.
func LoadTokenCounter(string) (TokenCounter, error) {
	return approxTokenCounter{}, nil
}
