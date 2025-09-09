// Package manifester defines core types used during manifest generation.
// These structs support future extensibility for variant metadata and manifest introspection.
package manifester

// ManifestMeta represents metadata about a single variant in the master manifest.
// Useful for debugging, analytics, or frontend introspection.
type ManifestMeta struct {
	Label       string // e.g. "720p_3000kbps"
	Bitrate     int    // e.g. 3000000 (in bits per second)
	Resolution  string // e.g. "1280x720"
	ManifestURL string // relative or absolute path to manifest
}
