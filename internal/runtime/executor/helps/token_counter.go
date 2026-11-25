package helps

// TokenCounter counts UTF-8 text. tiktoken codecs implement it.
// The notiktoken build tag substitutes a bytes-per-token approximation.
type TokenCounter interface {
	Count(text string) (int, error)
}
