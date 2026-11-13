package playlist

import "testing"

func TestParseSimplePlaylist(t *testing.T) {
	got := Parse("#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:7\n#EXTINF:2.5,\na.ts\n#EXTINF:3.0,\nb.ts\n")
	if len(got.Segments) != 2 || got.Segments[0].Sequence != 7 || got.Segments[1].StartMs != 2500 {
		t.Fatalf("unexpected playlist: %+v", got)
	}
}

func TestSpliceKeepsAbsoluteTimeline(t *testing.T) {
	text := "#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:40\n#EXT-X-PROGRAM-DATE-TIME:2026-11-13T10:00:00Z\n#EXTINF:2.5,\nseg40.ts\n#EXT-X-DISCONTINUITY\n#EXTINF:3.0,\nseg41.ts\n"
	got := Parse(text)
	if len(got.Segments) != 2 || got.Scanned != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
	first, second := got.Segments[0], got.Segments[1]
	if first.Sequence != 40 || second.Sequence != 41 {
		t.Fatalf("sequence must increment across discontinuity: %+v", got.Segments)
	}
	if first.ProgramTime != "2026-11-13T10:00:00Z" || second.ProgramTime != "2026-11-13T10:00:02.500Z" {
		t.Fatalf("absolute time not derived: %+v", got.Segments)
	}
	if second.StartMs != 2500 || !second.Discontinuity {
		t.Fatalf("discontinuity segment misplaced: %+v", second)
	}
}

func TestProgramTimeBindsOnlyToDecoratedSegment(t *testing.T) {
	got := Parse("#EXTM3U\n#EXT-X-PROGRAM-DATE-TIME:2026-11-13T10:00:00.000Z\n#EXTINF:1.25,\na.ts\n#EXTINF:2.5,\nb.ts\n")
	if got.Segments[1].ProgramTime != "2026-11-13T10:00:01.250Z" {
		t.Fatalf("derived time should follow duration: %+v", got.Segments)
	}
	// No program date time means no absolute time anywhere.
	plain := Parse("#EXTM3U\n#EXTINF:1.0,\na.ts\n#EXTINF:2.0,\nb.ts\n")
	for _, segment := range plain.Segments {
		if segment.ProgramTime != "" {
			t.Fatalf("unexpected program time: %+v", plain.Segments)
		}
	}
}

func TestTagsByteRangeBlankLinesAndUnknownTags(t *testing.T) {
	text := "#EXTM3U\n#EXT-X-TARGETDURATION:6\n#EXT-X-BYTERANGE:1000@0\n\n#EXTINF:2.0,\n\n#EXT-X-KEY:METHOD=AES-128,URI=\"k\"\na.ts\n#EXT-X-BYTERANGE:500\n#EXTINF:3.0,\nb.ts\n#EXT-X-ENDLIST\n"
	got := Parse(text)
	if len(got.Segments) != 2 || got.Scanned != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.Segments[0].URI != "a.ts" || got.Segments[1].URI != "b.ts" {
		t.Fatalf("URI order changed: %+v", got.Segments)
	}
	if got.Segments[0].DurationMs != 2000 || got.Segments[1].StartMs != 2000 {
		t.Fatalf("unexpected timeline: %+v", got.Segments)
	}
}

func TestMissingDurationAndEmptyInput(t *testing.T) {
	got := Parse("#EXTM3U\n#EXTINF:,\na.ts\n#EXTINF:1.5,\nb.ts\n")
	if len(got.Segments) != 2 || got.Segments[0].DurationMs != 0 || got.Segments[1].StartMs != 0 {
		t.Fatalf("missing duration should default to zero: %+v", got)
	}
	if empty := Parse(""); len(empty.Segments) != 0 || empty.Scanned != 0 {
		t.Fatalf("empty input: %+v", empty)
	}
}

func TestRepeatedMediaSequenceDoesNotRenumber(t *testing.T) {
	got := Parse("#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:5\n#EXTINF:1.0,\na.ts\n#EXT-X-MEDIA-SEQUENCE:5\n#EXTINF:1.0,\nb.ts\n")
	if got.Segments[0].Sequence != 5 || got.Segments[1].Sequence != 6 {
		t.Fatalf("sequence duplicated: %+v", got.Segments)
	}
}

func TestProgramTimeReanchorsAfterGap(t *testing.T) {
	got := Parse("#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:10\n#EXT-X-PROGRAM-DATE-TIME:2026-11-13T10:00:00Z\n#EXTINF:2.0,\na.ts\n#EXT-X-DISCONTINUITY\n#EXT-X-PROGRAM-DATE-TIME:2026-11-13T10:05:00Z\n#EXTINF:4.0,\nb.ts\n#EXTINF:1.0,\nc.ts\n")
	want := []string{"2026-11-13T10:00:00Z", "2026-11-13T10:05:00Z", "2026-11-13T10:05:04.000Z"}
	for i, w := range want {
		if got.Segments[i].ProgramTime != w {
			t.Fatalf("segment %d: want %s, got %+v", i, w, got.Segments)
		}
	}
	if got.Segments[1].Sequence != 11 || got.Segments[2].StartMs != 6000 {
		t.Fatalf("timeline or sequence disturbed: %+v", got.Segments)
	}
}
