//go:build nodiscovery

package discovery

import (
	"context"
	"fmt"
	"net"
)

// ZeroconfAdvertiser is the mDNS advertiser stub used when the nodiscovery build tag excludes libp2p/zeroconf.
type ZeroconfAdvertiser struct{}

// NewZeroconfAdvertiser reports that mDNS advertising was left out of this binary.
func NewZeroconfAdvertiser() *ZeroconfAdvertiser { return &ZeroconfAdvertiser{} }

// Start refuses to advertise because this binary was built without mDNS.
func (a *ZeroconfAdvertiser) Start(context.Context, ServiceSpec) error {
	return fmt.Errorf("mDNS discovery was excluded from this build (nodiscovery)")
}

// Stop is a no-op for the excluded advertiser.
func (a *ZeroconfAdvertiser) Stop() error { return nil }

// ZeroconfBrowser is the mDNS browser stub used when the nodiscovery build tag is set.
type ZeroconfBrowser struct{}

// NewZeroconfBrowser reports that LAN browsing was left out of this binary.
func NewZeroconfBrowser(...net.Interface) *ZeroconfBrowser { return &ZeroconfBrowser{} }

// Browse refuses to scan because this binary was built without mDNS.
func (b *ZeroconfBrowser) Browse(context.Context, string, string) ([]DiscoveredService, error) {
	return nil, fmt.Errorf("mDNS discovery was excluded from this build (nodiscovery)")
}

// BrowseWithFallback refuses to scan because this binary was built without mDNS.
func (b *ZeroconfBrowser) BrowseWithFallback(context.Context) ([]DiscoveredService, error) {
	return nil, fmt.Errorf("mDNS discovery was excluded from this build (nodiscovery)")
}

// BrowseWithFallbackServiceType refuses to scan because this binary was built without mDNS.
func (b *ZeroconfBrowser) BrowseWithFallbackServiceType(context.Context, string) ([]DiscoveredService, error) {
	return nil, fmt.Errorf("mDNS discovery was excluded from this build (nodiscovery)")
}
