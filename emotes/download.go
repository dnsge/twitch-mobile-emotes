package emotes

import (
	"bytes"
	"context"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	emoteSizeMap = map[ImageSize]int{
		ImageSizeSmall:  28,
		ImageSizeMedium: 56,
		ImageSizeLarge:  112,
	}
)

func DownloadEmote(emote Emote, size ImageSize) ([]byte, error) {
	url := emote.URL(size)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var img image.Image
	switch emote.Type() {
	case "png":
		pngImg, err := png.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
		img = pngImg
	case "gif":
		gifImg, err := gif.DecodeAll(resp.Body)
		if err != nil {
			return nil, err
		}
		img = selectGifFrame(emote, gifImg)
	default:
		return nil, fmt.Errorf("unsupported emote type %q", emote.Type())
	}

	return processImage(img, size)
}

func processImage(img image.Image, size ImageSize) ([]byte, error) {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	requiredSize := emoteSizeMap[size]

	var newImg image.Image
	if width != height {
		if width > height { // too wide, add top/bottom padding
			newImg = image.NewRGBA(image.Rect(0, 0, width, width))
			newImg = imaging.PasteCenter(newImg, img)
			newImg = imaging.Resize(newImg, requiredSize, requiredSize, imaging.Lanczos)
		} else { // too tall, add left/right padding
			newImg = image.NewRGBA(image.Rect(0, 0, height, height))
			newImg = imaging.PasteCenter(newImg, img)
			newImg = imaging.Resize(newImg, requiredSize, requiredSize, imaging.Lanczos)
		}
	} else {
		newImg = img
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, newImg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var (
	idealGifFrames = map[string]int{
		"b:5fa0159a40eb9502e2239963": 17, // modCheck
		"b:5df0742ee7df1277b6070125": 3,  // TriFi
		"b:5fa179c2710f8302f0c9e09e": 5,  // monkaX
		"b:5fa1f21e6f583802e38a7f29": 29, // WAYTOODANK
		"b:5fafdb852d853564472d6b70": 6,  // FeelsLateMan
	}
)

func selectGifFrame(emote Emote, g *gif.GIF) image.Image {
	key := emote.LetterCode() + ":" + emote.EmoteID()
	// frame number will be the found value or zero
	return getGifFrame(g, idealGifFrames[key])
}

func getGifFrame(g *gif.GIF, stopFrame int) image.Image {
	if stopFrame > len(g.Image)-1 || stopFrame < 0 { // safety checks
		stopFrame = 0
	}

	if len(g.Image) == 1 || stopFrame == 0 {
		return g.Image[0]
	}

	// Create canvas
	width, height := getGifDimensions(g)
	result := image.NewRGBA(image.Rect(0, 0, width, height))

	// Draw first frame
	draw.Draw(result, result.Bounds(), g.Image[0], image.Point{}, draw.Src)
	for i := 1; i <= stopFrame; i++ {
		// Yes, not a correct implementation of disposal methods but it works
		var op draw.Op
		switch g.Disposal[i-1] {
		case gif.DisposalBackground: // clear and replace existing with current frame
			op = draw.Src
		case gif.DisposalNone, gif.DisposalPrevious: // impose current frame over existing
			op = draw.Over
		}

		draw.Draw(result, result.Bounds(), g.Image[i], image.Point{}, op)
	}

	return result
}

func getGifDimensions(g *gif.GIF) (int, int) {
	var minX, maxX, minY, maxY int

	for _, img := range g.Image {
		if img.Rect.Min.X < minX {
			minX = img.Rect.Min.X
		}
		if img.Rect.Max.X > maxX {
			maxX = img.Rect.Max.X
		}
		if img.Rect.Min.Y < minY {
			minY = img.Rect.Min.Y
		}
		if img.Rect.Max.Y > maxY {
			maxY = img.Rect.Max.Y
		}
	}

	return maxX - minX, maxY - minY
}

type cacheMapValue struct {
	path    string
	created time.Time
}

func getFileKey(emote Emote, size ImageSize) string {
	return emote.LetterCode() + "_" + emote.EmoteID() + "_" + size.BttvString()
}

type ImageFileCache struct {
	cacheMap   map[string]cacheMapValue
	basePath   string
	expiration time.Duration
	// Remove files older than expiration in the basePath directory
	cleanOnIndex bool

	mu sync.Mutex
}

func NewImageFileCache(basePath string, expiration time.Duration, cleanOnIndex bool) *ImageFileCache {
	return &ImageFileCache{
		cacheMap:     make(map[string]cacheMapValue),
		basePath:     basePath,
		expiration:   expiration,
		cleanOnIndex: cleanOnIndex,
	}
}

func (c *ImageFileCache) Index() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := filepath.Walk(c.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".png" {
			key := strings.TrimSuffix(info.Name(), ".png")
			if time.Since(info.ModTime()) > c.expiration {
				if c.cleanOnIndex {
					if err := os.Remove(path); err != nil {
						return fmt.Errorf("remove %q: %w", key, err)
					}
				}
				// don't include file in index
				return nil
			}

			c.cacheMap[key] = cacheMapValue{
				path:    path,
				created: info.ModTime(),
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("cache index: %w", err)
	}
	return nil
}

func (c *ImageFileCache) Evict() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n := 0
	for key, val := range c.cacheMap {
		if time.Since(val.created) > c.expiration {
			if err := os.Remove(val.path); err != nil {
				return n, fmt.Errorf("remove %q: %w", key, err)
			}
			delete(c.cacheMap, key)
			n++
		}
	}
	return n, nil
}

func (c *ImageFileCache) Purge() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, val := range c.cacheMap {
		if err := os.Remove(val.path); err != nil {
			return fmt.Errorf("purge %q: %w", key, err)
		}
		delete(c.cacheMap, key)
	}

	return nil
}

func (c *ImageFileCache) AutoEvict(ctx context.Context) {
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if _, err := c.Evict(); err != nil {
				log.Printf("AutoEvict: %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *ImageFileCache) GetCachedOrDownload(emote Emote, size ImageSize, writer io.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getFileKey(emote, size)
	val, exists := c.cacheMap[key]
	if exists {
		if writer == nil {
			return nil
		}

		f, err := os.Open(val.path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(writer, f)
		return err
	} else {
		data, err := DownloadEmote(emote, size)
		if err != nil {
			return err
		}

		fPath := path.Join(c.basePath, key+".png")
		f, err := os.Create(fPath)
		if err != nil {
			return err
		}

		if _, err = f.Write(data); err != nil {
			return fmt.Errorf("write cache file: %w", err)
		}
		f.Close()

		c.cacheMap[key] = cacheMapValue{
			path:    fPath,
			created: time.Now(),
		}

		if writer != nil {
			_, err = writer.Write(data)
			return err
		} else {
			return nil
		}
	}
}
