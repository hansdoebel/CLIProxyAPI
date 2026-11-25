//go:build solaris || nowebrtc

package live

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

// The solaris build tag also covers illumos. Pion requires SO_REUSEPORT, which is unavailable on these targets.
// The nowebrtc tag omits the Pion stack from small builds such as the Raspberry Pi target.
func newMediaRelay(config.CodexLiveMediaRelayConfig, *mediaSessionLimiter) (mediaRelayFactory, error) {
	if runtime.GOOS == "solaris" || runtime.GOOS == "illumos" {
		return nil, fmt.Errorf("Codex Live WebRTC media relay on %s: %w", runtime.GOOS, errors.ErrUnsupported)
	}
	return nil, fmt.Errorf("Codex Live WebRTC media relay excluded by nowebrtc build tag: %w", errors.ErrUnsupported)
}
