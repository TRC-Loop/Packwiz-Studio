package widgets

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// RemoteImage shows an image fetched over HTTP, with a placeholder until
// it arrives. Mod icons come from Modrinth's CDN, so every card needs
// one of these.
//
// Loads are cached per URL for the life of the process: scrolling a
// result list back and forth would otherwise refetch every icon.
func RemoteImage(url string, size float32) fyne.CanvasObject {
	box := fyne.NewSize(size, size)
	slot := container.NewStack(placeholder(box))

	if url == "" {
		return slot
	}

	if res := cachedImage(url); res != nil {
		slot.Add(image(res, box))
		return slot
	}

	go func() {
		res := fetchImage(url)
		if res == nil {
			return
		}
		fyne.Do(func() {
			slot.Add(image(res, box))
			slot.Refresh()
		})
	}()

	return slot
}

// image builds a contained image at a fixed size.
func image(res fyne.Resource, box fyne.Size) *canvas.Image {
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(box)
	return img
}

// imageCache holds decoded icons by URL.
var imageCache struct {
	sync.Mutex
	entries map[string]fyne.Resource
}

// maxCachedImages bounds the cache. A search page holds twenty icons, so
// this covers a good deal of browsing before anything is dropped.
const maxCachedImages = 400

func cachedImage(url string) fyne.Resource {
	imageCache.Lock()
	defer imageCache.Unlock()
	return imageCache.entries[url]
}

func storeImage(url string, res fyne.Resource) {
	imageCache.Lock()
	defer imageCache.Unlock()

	if imageCache.entries == nil {
		imageCache.entries = map[string]fyne.Resource{}
	}
	// Once full the cache is cleared rather than evicting one entry at a
	// time. Icons are cheap to refetch and this keeps the bookkeeping to
	// nothing.
	if len(imageCache.entries) >= maxCachedImages {
		imageCache.entries = map[string]fyne.Resource{}
	}
	imageCache.entries[url] = res
}

// imageClient fetches icons. It is separate from the API client because
// icons come from a CDN rather than the API.
var imageClient = &http.Client{Timeout: 15 * time.Second}

// maxImageBytes caps one icon, so a mislabelled URL cannot pull a large
// file into memory.
const maxImageBytes = 4 << 20

// fetchImage downloads an icon. A failure is silent: the placeholder
// stays, which is a better outcome than an error dialog per missing icon.
func fetchImage(url string) fyne.Resource {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	resp, err := imageClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes))
	if err != nil {
		return nil
	}

	res := fyne.NewStaticResource(url, data)
	storeImage(url, res)
	return res
}
