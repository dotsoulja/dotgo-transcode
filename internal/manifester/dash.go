// Package manifester provides DASH master manifest generation.
// This file builds a multi-variant .mpd referencing segmented streams.
package manifester

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
)

// generateDASHMaster creates a basic DASH .mpd manifest referencing all variants.
// For simplicity, this assumes ffmpeg has already generated compliant segment sets.
//
// Output:
//
//	media/output/<slug>/master.mpd
//
// References:
//
//	<resolution>/<resolution>.mpd
func generateDASHMaster(seg *segmenter.SegmentResult) (string, error) {
	masterPath := filepath.Join(seg.OutputDir, "master.mpd")
	f, err := os.Create(masterPath)
	if err != nil {
		return "", NewManifesterError("write_file", "failed to create DASH master manifest", err)
	}
	defer f.Close()

	_, _ = f.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	_, _ = f.WriteString(`<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" minBufferTime="PT1.5S" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011">` + "\n")
	_, _ = f.WriteString(`  <Period>` + "\n")

	for _, manifest := range seg.Manifests {
		label := extractLabel(manifest)
		bitrate := estimateBitrate(label)

		// Reference manifest as <resolution>/<resolution>.mpd
		uri := filepath.Join(label, filepath.Base(manifest))

		_, _ = f.WriteString(fmt.Sprintf(
			`    <AdaptationSet mimeType="video/mp4" codecs="avc1.64001f" segmentAlignment="true" bitstreamSwitching="true">`+"\n"+
				`      <Representation id="%s" bandwidth="%d">`+"\n"+
				`        <BaseURL>%s</BaseURL>`+"\n"+
				`      </Representation>`+"\n"+
				`    </AdaptationSet>`+"\n",
			label, bitrate, uri,
		))
	}

	_, _ = f.WriteString(`  </Period>` + "\n")
	_, _ = f.WriteString(`</MPD>` + "\n")

	return masterPath, nil
}
