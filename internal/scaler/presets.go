// Package scaler provides resolution presets used for adaptive scaling.
// This file defines a canonical list of supported resolution presets.
package scaler

// StandardPresets defines a list of commonly used resolution presets.
// These are used as candidates during scaling decisions.
var StandardPresets = []ResolutionPreset{
	{
		Width:      3840,
		Height:     2160,
		Label:      "2160p",
		MinBitrate: 12000,
	},
	{
		Width:      2560,
		Height:     1440,
		Label:      "1440p",
		MinBitrate: 8000,
	},
	{
		Width:      1920,
		Height:     1080,
		Label:      "1080p",
		MinBitrate: 5000,
		IsDefault:  true,
	},
	{
		Width:      1280,
		Height:     720,
		Label:      "720p",
		MinBitrate: 2500,
	},
	{
		Width:      854,
		Height:     480,
		Label:      "480p",
		MinBitrate: 1000,
	},
	{
		Width:      640,
		Height:     360,
		Label:      "360p",
		MinBitrate: 600,
	},
	{
		Width:      426,
		Height:     240,
		Label:      "240p",
		MinBitrate: 300,
	},
}
