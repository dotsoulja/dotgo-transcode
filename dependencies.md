# dotgo-transcode Requirements

This module depends on the following external tools and system libraries:
    - This is not yet a full picture and is expected to evolve as development progresses

## System Dependencies

- **ffmpeg** - Required for transcoding, segmenting, and scaling
  - macOS-Install via Homebrew: `brew install ffmpeg`
  - Linux (apt): `sudo apt update`, `sudo apt install ffmpeg`

- **ffprobe** - Used for metadata inspection
  - Included with ffmpeg

## Optional Tools

- **Graphviz** - For generating architecture diagrams
- **ImageMagick** - _Reach: Used for thumbnail generation_

## Runtime notes

- Ensure `ffmpeg` and `ffprobe` are in your system `$PATH`
- Recommended minimum version: ffmpeg 5.0+


