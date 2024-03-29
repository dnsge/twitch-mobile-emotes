package emotes

import "fmt"

const (
	bttvCdnUrlFormat    = "https://cdn.betterttv.net/emote/%s/%s"
	ffzCdnUrlFormat     = "https://cdn.betterttv.net/frankerfacez_emote/%s/%s"
	sevenTVCDNUrlFormat = "https://cdn.7tv.app/emote/%s/%s"
)

type ImageSize int

const (
	ImageSizeSmall ImageSize = iota
	ImageSizeMedium
	ImageSizeLarge
)

func (s ImageSize) BttvString() string {
	switch s {
	case ImageSizeSmall:
		return "1x"
	case ImageSizeMedium:
		return "2x"
	case ImageSizeLarge:
		return "3x"
	default:
		panic("Unknown emote size")
	}
}

func (s ImageSize) FfzString() string {
	switch s {
	case ImageSizeSmall:
		return "1"
	case ImageSizeMedium:
		return "2"
	case ImageSizeLarge:
		return "4"
	default:
		panic("Unknown emote size")
	}
}

func (s ImageSize) SevenTVString() string {
	return s.FfzString() // use same behavior
}

func FormatBTTVEmote(id string, size ImageSize) string {
	return fmt.Sprintf(bttvCdnUrlFormat, id, size.BttvString())
}

func (b *BttvEmote) URL(size ImageSize) string {
	return FormatBTTVEmote(b.ID, size)
}

func FormatFFZEmote(id string, size ImageSize) string {
	return fmt.Sprintf(ffzCdnUrlFormat, id, size.FfzString())
}

func (f *FfzEmote) URL(size ImageSize) string {
	u := ""
	switch size {
	case ImageSizeSmall: // one -> two -> four
		if f.Images.One != "" {
			u = f.Images.One
		} else if f.Images.Two != "" {
			u = f.Images.Two
		} else {
			u = f.Images.Four
		}
	case ImageSizeMedium: // two -> one -> four
		if f.Images.Two != "" {
			u = f.Images.Two
		} else if f.Images.One != "" {
			u = f.Images.One
		} else {
			u = f.Images.Four
		}
	case ImageSizeLarge: // four -> two -> one
		if f.Images.Four != "" {
			u = f.Images.Four
		} else if f.Images.Two != "" {
			u = f.Images.Two
		} else {
			u = f.Images.One
		}
	default:
		panic("Unknown emote size")
	}

	return "https:" + u // FFZ image URLs don't have a schema attached
}

func (s *SevenTVEmote) URL(size ImageSize) string {
	expectedSizeID := size.SevenTVString()
	for _, sizeURLPair := range s.URLs {
		if sizeURLPair[0] == expectedSizeID {
			return sizeURLPair[1]
		}
	}

	// We didn't find it, build url based on blind luck? Will probably work.
	return fmt.Sprintf(sevenTVCDNUrlFormat, s.ID, expectedSizeID+"x")
}
