package playlist

import (
	"bufio"
	"strconv"
	"strings"
)

type Segment struct {
	Sequence      int64  `json:"sequence"`
	URI           string `json:"uri"`
	DurationMs    int64  `json:"duration_ms"`
	StartMs       int64  `json:"start_ms"`
	Discontinuity bool   `json:"discontinuity"`
	ProgramTime   string `json:"program_time,omitempty"`
}

type Result struct {
	Segments []Segment `json:"segments"`
	Scanned  int       `json:"scanned"`
}

func Parse(text string) Result {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var seq int64
	var duration int64
	var start int64
	var programTime string
	var discontinuity bool
	segments := make([]Segment, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			seq, _ = strconv.ParseInt(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"), 10, 64)
		case strings.HasPrefix(line, "#EXTINF:"):
			value := strings.TrimSuffix(strings.TrimPrefix(line, "#EXTINF:"), ",")
			seconds, _ := strconv.ParseFloat(value, 64)
			duration = int64(seconds * 1000)
		case strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:"):
			programTime = strings.TrimPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:")
		case line == "#EXT-X-DISCONTINUITY":
			discontinuity = true
		case line != "" && !strings.HasPrefix(line, "#"):
			segments = append(segments, Segment{Sequence: seq, URI: line, DurationMs: duration, StartMs: start, Discontinuity: discontinuity, ProgramTime: programTime})
			seq++
			start += duration
			duration = 0
			programTime = ""
			discontinuity = false
		}
	}
	return Result{Segments: segments, Scanned: len(segments)}
}
