package emotes

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	emoteSizeMap = map[ImageSize]int{
		ImageSizeSmall:  28,
		ImageSizeMedium: 56,
		ImageSizeLarge:  112,
	}
)

var idealGifFrames map[string]int = nil

func InitIdealGifFrames(path string) {
	idealGifFrames = make(map[string]int)

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Open ideal gif file: %v\n", err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Read ideal gif file: %v\n", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	for n, l := range lines {
		if len(l) == 0 || l[0] == '#' { // empty line or comment
			continue
		}

		if comment := strings.Index(l, "#"); comment != -1 {
			l = l[:comment]
		}

		l = strings.Trim(l, " \t\r")
		parts := strings.Split(l, ":")
		if len(parts) != 3 {
			log.Printf("Warning: invalid ideal gif encoding on line %d\n", n)
			continue
		}

		val, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("Warning: invalid interger value on line %d %q\n", n, parts[2])
			continue
		}

		idealGifFrames[parts[0]+":"+parts[1]] = val
	}
	fmt.Println(idealGifFrames)
}

func processImage(img image.Image, size ImageSize) ([]byte, error) {
	requiredSize := emoteSizeMap[size]
	newImg := resizeImageWithAspectRatio(img, requiredSize, requiredSize)

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, newImg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func processImageHalves(img image.Image, size ImageSize) ([]byte, []byte, error) {
	requiredSize := emoteSizeMap[size]

	newImg := resizeImageWithAspectRatio(img, requiredSize * 2, requiredSize) // double width for wide emotes
	leftImg := imaging.Crop(newImg, image.Rect(0, 0, requiredSize, requiredSize))
	rightImg := imaging.Crop(newImg, image.Rect(requiredSize, 0, requiredSize * 2, requiredSize))

	left := new(bytes.Buffer)
	right := new(bytes.Buffer)

	if err := png.Encode(left, leftImg); err != nil {
		return nil, nil, fmt.Errorf("encode left half: %w", err)
	}
	if err := png.Encode(right, rightImg); err != nil {
		return nil, nil, fmt.Errorf("encode right half: %w", err)
	}

	return left.Bytes(), right.Bytes(), nil
}

func resizeImageWithAspectRatio(img image.Image, targetWidth, targetHeight int) image.Image {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	ratio := float64(width) / float64(height)
	targetRatio := float64(targetWidth) / float64(targetHeight)

	fmt.Printf("w: %d h: %d r: %.2f tr: %.2f\n", width, height, ratio, targetRatio)

	var newImg image.Image
	if targetRatio > ratio { // target is wider, scale to height first then resize
		newImg = image.NewRGBA(image.Rect(0, 0, int(targetRatio * float64(height)), height))
	} else if targetRatio < ratio { // target is taller, scale to width first then resize
		newImg = image.NewRGBA(image.Rect(0, 0, width, int(float64(width) / targetRatio)))
	} else {
		return imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)
	}
	newImg = imaging.PasteCenter(newImg, img)
	newImg = imaging.Resize(newImg, targetWidth, targetHeight, imaging.Lanczos)
	return newImg
}

func selectGifFrame(emote Emote, g *gif.GIF) image.Image {
	key := emote.LetterCode() + ":" + emote.EmoteID()
	if idealGifFrames == nil {
		return getGifFrame(g, 0)
	} else {
		// frame number will be the found value or zero
		return getGifFrame(g, idealGifFrames[key])
	}
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
