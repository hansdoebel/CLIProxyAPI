package cache

// Runtime cache caps default to the exported constants. ApplyLowResourceLimits
// tightens them for small machines. Tests keep using the constants, which stay
// the default until this function runs.
var (
	runtimeCodexReplayMaxEntries       int
	runtimeCodexReplayMaxBytes         int
	runtimeAntigravityReplayMaxEntries int
	runtimeAntigravityReplayMaxBytes   int
	runtimeClaudeReplayMaxEntries      int
	runtimeClaudeReplayMaxTotalBytes   int
	runtimeKimiReplayMaxEntries        int
	runtimeKimiReplayMaxBytes          int
	runtimeKimiReplayMaxTotalBytes     int
	runtimeXAIReplayMaxEntries         int
	runtimeSignatureMaxEntriesPerGroup int
)

func codexReplayMaxEntries() int {
	if runtimeCodexReplayMaxEntries > 0 {
		return runtimeCodexReplayMaxEntries
	}
	return CodexReasoningReplayCacheMaxEntries
}

func codexReplayMaxBytesPerEntry() int {
	if runtimeCodexReplayMaxBytes > 0 {
		return runtimeCodexReplayMaxBytes
	}
	return CodexReasoningReplayCacheMaxBytesPerEntry
}

func antigravityReplayMaxEntries() int {
	if runtimeAntigravityReplayMaxEntries > 0 {
		return runtimeAntigravityReplayMaxEntries
	}
	return AntigravityReasoningReplayCacheMaxEntries
}

func antigravityReplayMaxBytesPerEntry() int {
	if runtimeAntigravityReplayMaxBytes > 0 {
		return runtimeAntigravityReplayMaxBytes
	}
	return AntigravityReasoningReplayCacheMaxBytesPerEntry
}

func claudeReplayMaxEntries() int {
	if runtimeClaudeReplayMaxEntries > 0 {
		return runtimeClaudeReplayMaxEntries
	}
	return ClaudeThinkingReplayCacheMaxEntries
}

func claudeReplayMaxTotalBytes() int {
	if runtimeClaudeReplayMaxTotalBytes > 0 {
		return runtimeClaudeReplayMaxTotalBytes
	}
	return ClaudeThinkingReplayCacheMaxTotalBytes
}

func kimiReplayMaxEntries() int {
	if runtimeKimiReplayMaxEntries > 0 {
		return runtimeKimiReplayMaxEntries
	}
	return KimiThinkingReplayCacheMaxEntries
}

func kimiReplayMaxBytesPerEntry() int {
	if runtimeKimiReplayMaxBytes > 0 {
		return runtimeKimiReplayMaxBytes
	}
	return KimiThinkingReplayCacheMaxBytesPerEntry
}

func kimiReplayMaxTotalBytes() int {
	if runtimeKimiReplayMaxTotalBytes > 0 {
		return runtimeKimiReplayMaxTotalBytes
	}
	return KimiThinkingReplayCacheMaxTotalBytes
}

func xaiReplayMaxEntries() int {
	if runtimeXAIReplayMaxEntries > 0 {
		return runtimeXAIReplayMaxEntries
	}
	return XAIReasoningReplayCacheMaxEntries
}

// ApplyLowResourceLimits caps in-memory replay and signature caches for a
// single-user device. It is safe to call more than once.
func ApplyLowResourceLimits() {
	const (
		maxEntries = 256
		maxBytes   = 2 << 20
		maxTotal   = 8 << 20
	)
	runtimeCodexReplayMaxEntries = maxEntries
	runtimeCodexReplayMaxBytes = maxBytes
	runtimeAntigravityReplayMaxEntries = maxEntries
	runtimeAntigravityReplayMaxBytes = maxBytes
	runtimeClaudeReplayMaxEntries = maxEntries
	runtimeClaudeReplayMaxTotalBytes = maxTotal
	runtimeKimiReplayMaxEntries = maxEntries
	runtimeKimiReplayMaxBytes = maxBytes
	runtimeKimiReplayMaxTotalBytes = maxTotal
	runtimeXAIReplayMaxEntries = maxEntries
	runtimeSignatureMaxEntriesPerGroup = 128
}

func trimSignatureGroup(sc *groupCache) {
	maxEntries := runtimeSignatureMaxEntriesPerGroup
	if maxEntries <= 0 || sc == nil || len(sc.entries) <= maxEntries {
		return
	}
	for len(sc.entries) > maxEntries {
		var oldestKey string
		var found bool
		for key, entry := range sc.entries {
			if !found || entry.Timestamp.Before(sc.entries[oldestKey].Timestamp) {
				oldestKey = key
				found = true
			}
		}
		if !found {
			return
		}
		delete(sc.entries, oldestKey)
	}
}
