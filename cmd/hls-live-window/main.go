package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/hlswindow/playlist"
)

func main() {
	sample := flag.String("sample", "splice", "sample name")
	flag.Parse()
	var text string
	switch *sample {
	case "splice":
		text = "#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:40\n#EXT-X-PROGRAM-DATE-TIME:2026-11-13T10:00:00Z\n#EXTINF:2.5,\nseg40.ts\n#EXT-X-DISCONTINUITY\n#EXTINF:3.0,\nseg41.ts\n"
	case "live":
		text = "#EXTM3U\n#EXT-X-MEDIA-SEQUENCE:100\n#EXTINF:6.0,\na.ts\n#EXTINF:6.0,\nb.ts\n"
	default:
		fmt.Fprintln(os.Stderr, "unknown sample")
		os.Exit(2)
	}
	data, _ := json.Marshal(playlist.Parse(text))
	fmt.Println(string(data))
}
