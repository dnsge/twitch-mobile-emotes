package emotes

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
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

func requestEmote(emote Emote, size ImageSize) (image.Image, error) {
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

	return img, nil
}

func DownloadEmote(emote Emote, size ImageSize) ([]byte, error) {
	img, err := requestEmote(emote, size)
	if err != nil {
		return nil, fmt.Errorf("request emote: %w", err)
	}
	return processImage(img, size)
}

func DownloadEmoteHalves(emote Emote, size ImageSize) ([]byte, []byte, error) {
	img, err := requestEmote(emote, size)
	if err != nil {
		return nil, nil, fmt.Errorf("request emote: %w", err)
	}
	l, r, err := processImageHalves(img, size)
	if err != nil {
		return nil, nil, fmt.Errorf("process halves: %w", err)
	}
	return l, r, nil
}

type cacheMapValue struct {
	path    string
	created time.Time
}

func hashString(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func getFileKey(emote Emote, size ImageSize) string {
	return emote.LetterCode() + "_" + emote.EmoteID() + "_" + size.BttvString()
}

func getVirtualFileKey(emote Emote, size ImageSize, half VirtualHalf) string {
	return "v" + half.LetterCode() + "_" + emote.LetterCode() + "_" + emote.EmoteID() + "_" + size.BttvString()
}

func getAspectRatioKey(emote Emote) string {
	return hashString(emote.LetterCode() + "_" + emote.EmoteID())
}

type ImageFileCache struct {
	cacheMap       map[string]cacheMapValue
	aspectRatioMap map[string]float64
	basePath       string
	expiration     time.Duration
	// Remove files older than expiration in the basePath directory
	cleanOnIndex bool

	mu sync.Mutex
}

func NewImageFileCache(basePath string, expiration time.Duration, cleanOnIndex bool) *ImageFileCache {
	return &ImageFileCache{
		cacheMap:       make(map[string]cacheMapValue),
		aspectRatioMap: make(map[string]float64),
		basePath:       basePath,
		expiration:     expiration,
		cleanOnIndex:   cleanOnIndex,
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

		if err := c.writeDataToCache(key, data); err != nil {
			return err
		}

		if writer != nil {
			_, err = writer.Write(data)
			return err
		} else {
			return nil
		}
	}
}

func (c *ImageFileCache) DownloadToCache(emote Emote, size ImageSize) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getFileKey(emote, size)
	_, exists := c.cacheMap[key]
	if !exists {
		data, err := DownloadEmote(emote, size)
		if err != nil {
			return err
		}

		if err := c.writeDataToCache(key, data); err != nil {
			return err
		}
	}
	return nil
}

type VirtualHalf int

const (
	LeftHalf VirtualHalf = iota
	RightHalf
)

func (h VirtualHalf) LetterCode() string {
	switch h {
	case LeftHalf:
		return "l"
	case RightHalf:
		return "r"
	default:
		return "?"
	}
}

func (c *ImageFileCache) GetCachedOrDownloadHalf(emote Emote, size ImageSize, half VirtualHalf, writer io.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getVirtualFileKey(emote, size, half)
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
		left, right, err := DownloadEmoteHalves(emote, size)
		if err != nil {
			return err
		}

		leftKey := getVirtualFileKey(emote, size, LeftHalf)
		rightKey := getVirtualFileKey(emote, size, RightHalf)

		if err := c.writeDataToCache(leftKey, left); err != nil {
			return err
		}
		if err := c.writeDataToCache(rightKey, right); err != nil {
			return err
		}

		if writer != nil {
			var data []byte
			switch half {
			case LeftHalf:
				data = left
			case RightHalf:
				data = right
			default:
				return fmt.Errorf("unknown half %d", half)
			}

			_, err = writer.Write(data)
			return err
		} else {
			return nil
		}
	}
}

func (c *ImageFileCache) DownloadVirtualToCache(emote Emote, size ImageSize) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	leftKey := getVirtualFileKey(emote, size, LeftHalf)
	rightKey := getVirtualFileKey(emote, size, RightHalf)
	_, existLeft := c.cacheMap[leftKey]
	_, existRight := c.cacheMap[rightKey]
	if !existLeft || !existRight {
		left, right, err := DownloadEmoteHalves(emote, size)
		if err != nil {
			return err
		}

		if err := c.writeDataToCache(leftKey, left); err != nil {
			return err
		}
		if err := c.writeDataToCache(rightKey, right); err != nil {
			return err
		}
	}

	return nil
}

func (c *ImageFileCache) writeDataToCache(key string, data []byte) error {
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
	return nil
}

func (c *ImageFileCache) GetEmoteAspectRatio(emote Emote) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getAspectRatioKey(emote)
	val, found := c.aspectRatioMap[key]
	if found {
		return val, nil
	}

	url := emote.URL(ImageSizeSmall)
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var cfg image.Config
	if emote.Type() == "png" {
		c, err := png.DecodeConfig(resp.Body)
		if err != nil {
			return 0, err
		}
		cfg = c
	} else if emote.Type() == "gif" {
		c, err := gif.DecodeConfig(resp.Body)
		if err != nil {
			return 0, err
		}
		cfg = c
	} else {
		return 0, fmt.Errorf("unsupported emote type %q", emote.Type())
	}

	calculated := float64(cfg.Width) / float64(cfg.Height)

	c.aspectRatioMap[key] = calculated
	return calculated, nil
}
