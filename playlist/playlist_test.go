package playlist

import "testing"

func TestParseSimplePlaylist(t *testing.T) {
	got := Parse("#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:7\n#EXTINF:2.5,\na.ts\n#EXTINF:3.0,\nb.ts\n")
	if len(got.Segments) != 2 || got.Segments[0].Sequence != 7 || got.Segments[1].StartMs != 2500 {
		t.Fatalf("unexpected playlist: %+v", got)
	}
}
