package transcoder

// DefaultVariants defines a standard set of video variants.
// These can be used when no custom variants are specified.
// Each entry includes resolution - bitrate pairs in ffmpeg-compatible format.
var DefaultVariants = []Variant{
	{Resolution: "1080p", Bitrate: "8000k"},
	{Resolution: "1080p", Bitrate: "5000k"},
	{Resolution: "720p", Bitrate: "3000k"},
	{Resolution: "720p", Bitrate: "2500k"},
	{Resolution: "480p", Bitrate: "1500k"},
	{Resolution: "360p", Bitrate: "1000k"},
	{Resolution: "240p", Bitrate: "500k"},
	{Resolution: "144p", Bitrate: "150k"},
}
