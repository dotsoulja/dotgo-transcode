# dotgo-transcode Roadmap

A modular Go library for dynamic media transcoding and adaptive streaming. This roadmap outlines the development phases, goals, and rationale behind each component. It serves both as a personal guide a public-facing blueprint.

---

## Phase 1: Project Initialization

### What
Set up the foundational structure of the project: Git repo, Go module, directory layout, and initial documentation.

### Why
A clean, modular architecture ensures scalability, clarity, and ease of integration with other systems, (current media server)

### How
- Initialize Git and Go module:
  ```bash
  git init
  go mod init github.com/dotsoulja/dotgo-transcode
  ```
- Create directory structure:
  ```bash
  mkdir -p cmd internal/{analyzer,transcoder,segmenter,scaler,muxer} profiles examples docs tests
  ```
- Draft `README.md` with project goals, usage scenarios, and roadmap link.
- Add `.gitignore` and `LICENSE`.

---

## Phase 2: Define TranscodeProfile

### What
Create a struct that defines transcoding parameters: input resolution, target resolutions, codecs, bitrate ranges, etc.

### Why
This will become the central configuration object for all transcoding operations: modular, reusable, and easy to serialize.

### How
- Define `TranscodeProfile` in `internal/transcoder/profile.go`
- Support JSON/YAML loading from `profiles/`
- Include fields like:
  ```go
  type TranscodeProfile struct {
        InputPath     string
        OutputDir     string
        TargetRes     []string          // e.g. ["1080p", "720p", "480p"]
        Codec         string
        Bitrate       map[string]string
        SegmentLength int
        Container     string
  }
  ```

---

## Phase 3: Analyzer Module

### What
Inspect input media files to extract metadata: resolution, duration, codecs, keyframes, etc.

### Why
Understanding the source file is essential before deciding how to transcode it. This will also help with fallback logic and error handling.

### How
- Initial: Wrap `ffprobe` Reach: use Go bindings, to extract metadata
- Output a `MediaInfo` struct with fields like:
  ```go
  type MediaInfo struct {
      Width     int
      Height    int
      Duration  float64
      Codec     string
      Bitrate   int
      Keyframes []float64
  }
  ```
- Log results for forensic clarity.

---

## Phase 4: Transcoder Core

### What
Orchestrate the actual transcoding process using either ffmpeg or a Go-native wrapper.

### Why
This is the heart of the module where input becomes output, and resolution variants are generated.

### How
- Accept `TranscodeProfile` and `MediaInfo`
- Spawn subprocesses or use bindings to run ffmpeg commands
- Ensure:
  - Keyframe alignment
  - Codec compatibility
  - Segment length consistency
- Log each step with timestamps and exit codes

---

## Phase 5: Scaler Logic

### What
Handle resolution scaling-both automatic (based on network/ client cues) and manual (user-selected).

### Why
Adaptive streaming depends on having multiple resolution variants ready to serve.

### How
- Define resolution presets (1080p, 720p, 480p, etc)
- Implement scaling logic:
  - Downscale from source
  - Skip if source is lower than target
- Add hooks for client-side override or network-based fallback

---

## Phase 6: Segmenter and Manifest Generator

### What
Split media into segments (HLS/DASH) and generate manifest files (.m3u8, .mpd).

### Why
Segmented streaming allows for smooth playback, resolution switching, and better buffering.


### How
- Use ffmpeg to generate `.ts` or `.mp4` segments
- Align segments across resolutions
- Generate manifests with variant playlists
- Include metadata for client-side switching

---

## Phase 7: Testing and Validation

### What
Write unit and integration tests to validate each module

### Why
Ensure reliability, catch regressions, and document edge cases.

### How
- Use Go's `testing` package
- Create mock profiles and media files
- Validate:
  - Metadata extraction
  - Transcoding output
  - Segment alignment
  - Manifest correctness

---

## Phase 8: Integration Examples

### What
Show how to plug the module into your media server or other systems.

### Why
Demonstrates real-world usage and helps others adopt the library.

### How
- Create sample CLI in `cmd/`
- Add example server hooks in `examples/`
- Document API calls, expected inputs/outputs

---

## Phase 9: Documentation and Developer Experience

### What
Write guides, inline comments, and edge-case explanations.

### Why
Make the library intuitive, shareable, and developer-friendly.

### How
- Add `docs/architecture.md`, `docs/usage.md`, `docs/edge-cases.md`
- Include diagrams and flowcharts as needed
- Document error codes and recovery strategies

---

## Phase 10: Polish and Release

### What
Refine, tag, and publish the library.

### Why
Make it ready for public use or private deployment.

### How
- Add versioning and changelog
- Tag release with semantic versioning
- Optionally publish pkg.go.dev

---

