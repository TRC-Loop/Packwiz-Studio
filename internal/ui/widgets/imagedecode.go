package widgets

import (
	"bytes"
	"image"
	"image/png"
	"strings"

	"fyne.io/fyne/v2"

	// Registering these decoders is what lets image.Decode recognise a
	// downloaded icon. Modrinth serves most of its icons as WebP, which
	// neither the standard library nor Fyne handles on its own, so an
	// undecoded one reaches the canvas and fails to render on every draw.
	_ "image/gif"
	_ "image/jpeg"

	_ "golang.org/x/image/webp"
)

// decodeImage turns downloaded bytes into a resource Fyne can draw.
//
// It returns false for anything undecodable rather than handing the
// canvas a resource it will fail on: a broken image logs an error every
// time it is drawn, which is once per frame.
func decodeImage(name string, data []byte) (fyne.Resource, bool) {
	if len(data) == 0 {
		return nil, false
	}

	// Fyne renders SVG itself, decided by the resource name's extension,
	// so an SVG is passed through untouched under a name that says so.
	if looksLikeSVG(data) {
		return fyne.NewStaticResource(resourceName(name, ".svg"), data), true
	}

	decoded, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, false
	}

	// Formats Fyne draws natively are passed through as they arrived,
	// which avoids re-encoding the common case.
	switch format {
	case "png", "jpeg", "gif":
		return fyne.NewStaticResource(resourceName(name, "."+format), data), true
	}

	// Anything else, WebP in practice, is re-encoded to PNG.
	var buf bytes.Buffer
	if err := png.Encode(&buf, decoded); err != nil {
		return nil, false
	}
	return fyne.NewStaticResource(resourceName(name, ".png"), buf.Bytes()), true
}

// looksLikeSVG sniffs for an SVG document, which has no magic number.
func looksLikeSVG(data []byte) bool {
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	lower := strings.ToLower(string(head))

	return strings.Contains(lower, "<svg") ||
		(strings.Contains(lower, "<?xml") && strings.Contains(lower, "svg"))
}

// resourceName gives a resource a name ending in ext, because Fyne
// decides how to draw a static resource from its extension.
func resourceName(name, ext string) string {
	trimmed := name
	if i := strings.LastIndexAny(trimmed, "/"); i >= 0 {
		trimmed = trimmed[i+1:]
	}
	if i := strings.Index(trimmed, "?"); i >= 0 {
		trimmed = trimmed[:i]
	}
	if trimmed == "" {
		trimmed = "image"
	}

	if strings.HasSuffix(strings.ToLower(trimmed), ext) {
		return trimmed
	}
	if i := strings.LastIndex(trimmed, "."); i > 0 {
		trimmed = trimmed[:i]
	}
	return trimmed + ext
}
