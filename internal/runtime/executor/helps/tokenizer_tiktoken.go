//go:build !notiktoken

package helps

import "github.com/tiktoken-go/tokenizer"

// LoadTokenCounter returns a tiktoken codec. kind selects the encoding used by
// the default build. Unknown kinds fall back to o200k_base.
func LoadTokenCounter(kind string) (TokenCounter, error) {
	switch kind {
	case "cl100k":
		return tokenizer.Get(tokenizer.Cl100kBase)
	case "o200k":
		return tokenizer.Get(tokenizer.O200kBase)
	case "gpt-5":
		return tokenizer.ForModel(tokenizer.GPT5)
	case "gpt-4.1":
		return tokenizer.ForModel(tokenizer.GPT41)
	case "gpt-4o":
		return tokenizer.ForModel(tokenizer.GPT4o)
	case "gpt-4":
		return tokenizer.ForModel(tokenizer.GPT4)
	case "gpt-3.5":
		return tokenizer.ForModel(tokenizer.GPT35Turbo)
	case "o1":
		return tokenizer.ForModel(tokenizer.O1)
	case "o3":
		return tokenizer.ForModel(tokenizer.O3)
	case "o4-mini":
		return tokenizer.ForModel(tokenizer.O4Mini)
	default:
		return tokenizer.Get(tokenizer.O200kBase)
	}
}
