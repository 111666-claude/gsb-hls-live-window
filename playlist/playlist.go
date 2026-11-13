package playlist

import (
	"bufio"
	"math"
	"strconv"
	"strings"
	"time"
)

// programTimeLayout renders a derived wall-clock time as UTC RFC 3339 with
// millisecond precision, e.g. 2026-11-13T10:00:02.500Z.
const programTimeLayout = "2006-01-02T15:04:05.000Z"

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

// parser is a single-pass state machine over the playlist text. The timeline
// cursor (timelineMs) is the start of the next segment and is never reset, so
// discontinuities only switch encoding era and do not move absolute time.
type parser struct {
	segments []Segment

	sequence int64

	timelineMs int64

	durationMs int64

	pendingTime *time.Time
	pendingText string

	discontinuity bool

	// Absolute-time anchor: anchorTime is the wall clock of the segment
	// starting at anchorStart.
	anchorTime  time.Time
	anchorStart int64
	hasAnchor   bool
}

func Parse(text string) Result {
	// A UTF-8 BOM is not whitespace, so drop it before line handling.
	text = strings.TrimPrefix(text, "\ufeff")

	p := &parser{segments: make([]Segment, 0)}
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			p.tag(line)
			continue
		}
		p.segment(line)
	}
	return Result{Segments: p.segments, Scanned: len(p.segments)}
}

func (p *parser) tag(line string) {
	switch {
	case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
		// The tag declares the sequence of the first segment in this window.
		// Once segments have been parsed the window is already numbered, so
		// a later tag must not renumber it and create duplicates.
		if len(p.segments) > 0 {
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"))
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
			p.sequence = value
		}

	case strings.HasPrefix(line, "#EXTINF:"):
		raw := strings.TrimPrefix(line, "#EXTINF:")
		if comma := strings.IndexByte(raw, ','); comma >= 0 {
			raw = raw[:comma]
		}
		raw = strings.TrimSpace(raw)

		// A missing or malformed duration defaults to zero; the URI is still
		// part of the playlist and keeps its position.
		p.durationMs = 0
		if raw == "" {
			return
		}
		seconds, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 {
			return
		}
		p.durationMs = int64(math.Round(seconds * 1000))

	case strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:"):
		raw := strings.TrimSpace(strings.TrimPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:"))
		value, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			// Leave the pending anchor untouched; an unparseable tag cannot
			// anchor wall-clock time.
			return
		}
		// Only the next segment is decorated: a pending tag never leaks to
		// following segments, which are derived from the timeline instead.
		p.pendingTime = &value
		p.pendingText = raw

	case line == "#EXT-X-DISCONTINUITY":
		p.discontinuity = true

	default:
		// EXT-X-BYTERANGE, EXT-X-KEY, EXT-X-MAP, EXT-X-ENDLIST and any other
		// unknown tag carry no timeline information and are skipped.
	}
}

func (p *parser) segment(uri string) {
	start := p.timelineMs
	programTime := ""

	if p.pendingTime != nil {
		// A program date time binds to exactly the segment it decorates and
		// re-anchors the absolute timeline at that segment.
		programTime = p.pendingText
		p.anchorTime = *p.pendingTime
		p.anchorStart = start
		p.hasAnchor = true
	} else if p.hasAnchor {
		// Later segments inherit absolute time purely from the continuous
		// timeline, including across discontinuities.
		offset := start - p.anchorStart
		programTime = p.anchorTime.Add(time.Duration(offset) * time.Millisecond).
			UTC().Format(programTimeLayout)
	}

	p.segments = append(p.segments, Segment{
		Sequence:      p.sequence,
		URI:           uri,
		DurationMs:    p.durationMs,
		StartMs:       start,
		Discontinuity: p.discontinuity,
		ProgramTime:   programTime,
	})

	// Sequence increments once per segment and is unaffected by
	// discontinuities; start_ms is contiguous with the previous segment.
	p.sequence++
	p.timelineMs += p.durationMs

	p.durationMs = 0
	p.pendingTime = nil
	p.pendingText = ""
	p.discontinuity = false
}
