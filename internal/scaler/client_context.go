// Package scaler defines optional client-side context for scaling decisions.
// This file contains hints about playback environment, device type, and network conditions.
package scaler

// ClientContext represents the playback environment and client capabilities.
// It is used to guide resolution selection and adaptive scaling.
type ClientContext struct {
	DeviceType      string // e.g. "mobile", "desktop", "tv"
	BandwidthKbps   int    // Current estimated bandwidth in Kbps
	PreferUpscale   bool   // If true, prefers higher resolution even if bandwidth is borderline
	AllowLowRes     bool   // If false, restricts resolution below a certain threshold
	ManualOverride  string // If set, forces a specific resolution (e.g. "720p")
	RecentFailures  int    // Number of recent playback stalls or buffering events
	AdaptiveEnabled bool   // Enables dynamic resolution switching
}

// IsMobile returns true if the device is mobile
func (c *ClientContext) IsMobile() bool {
	return c != nil && c.DeviceType == "mobile"
}

// IsBandwidthConstrained returns true if bandwidth is below HD threshold
func (c *ClientContext) IsBandwidthConstrained() bool {
	return c != nil && c.BandwidthKbps > 0 && c.BandwidthKbps < 2500
}
