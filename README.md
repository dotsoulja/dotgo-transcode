# dotgo-transcode

A modular, standalone Go library for dynamic media transcoding and adaptive streaming. Designed with the intent to integrate with a self-hosted media server, but fully independent and shareable.

## Planned Features
- Dynamic resolution scaling, both automatic and manual (e.g. adjusting by network strength and/or accepting user requests).
- Segment-based streaming (HLS/DASH)
- Client-aware fallback logic
- Modular architecture with clean interfaces promoting ease of use
- Inline docs and potential edge-case guides

## Usage Scenarios
- Self-hosted media server with adaptive playback
- CDN-style segment generation
- Real-time resolution switching based on network conditions
    - Stretch: integrate features to allow other environmental conditions to update resolutions updates

## Getting Started
```bash
go get github.com/dotsoulja/dotgo-transcode
