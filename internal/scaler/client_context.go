// Package scaler defines optional client-side context for scaling decisions.
// This file contains hints about playback environment, device type, and network conditions.
package scaler

// ClientContext provides environmental cues for resolution selection.
// This struct is optional but enables adaptive logic based on client capabilities.
type ClientContext struct {
	DeviceType    string // e.g. "mobile", "desktop", "tv"
	BandwidthKbps int    // Estimated available bandwidth in kbps
	PreferUpscale bool   // Whether the client prefers upscaling over downscaling
	AllowLowRes   bool   // Whether the client accepts resolutions below 480p
}

// IsMobile returns true if the device is mobile
func (c *ClientContext) IsMobile() bool {
	return c != nil && c.DeviceType == "mobile"
}

// IsBandwidthConstrained returns true if bandwidth is below HD threshold
func (c *ClientContext) IsBandwidthConstrained() bool {
	return c != nil && c.BandwidthKbps > 0 && c.BandwidthKbps < 2500
}
