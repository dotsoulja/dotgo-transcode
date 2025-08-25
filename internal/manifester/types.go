// Package manifester defines core types used during manifest generation.
// These structs support future extensibility for variant metadata and manifest introspection.
package manifester

// ManifestMeta represents metadata about a single variant in the master manifest.
// Useful for debugging, analytics, or frontend introspection.
type ManifestMeta struct {
	Label       string // e.g. "720p"
	Bitrate     int    // e.g. 1500
	Resolution  string // e.g. "1280x720"
	ManifestURL string // relative or absolute path to manifest
}
