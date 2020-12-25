package tme

import (
	"fmt"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"log"
	"net/http"
	"strings"
)

func getSizeFromString(text string) (emotes.ImageSize, error) {
	switch text {
	case "1.0":
		return emotes.ImageSizeSmall, nil
	case "2.0":
		return emotes.ImageSizeMedium, nil
	case "4.0":
		return emotes.ImageSizeLarge, nil
	default:
		return -1, fmt.Errorf("Unknown image size %s\n", text)
	}
}

func handleEmoticonRequest(w http.ResponseWriter, r *http.Request, store *emotes.EmoteStore, cache *emotes.ImageFileCache, highRes bool) {
	// URL is in format of "/emoticons/v1/<id>/<size>
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 5 || parts[1] != "emoticons" || parts[2] != "v1" { // verify URL
		http.NotFound(w, r)
		return
	}

	id := parts[3]

	var size emotes.ImageSize
	if highRes {
		size = emotes.ImageSizeLarge
	} else {
		s, err := getSizeFromString(parts[4])
		if err != nil {
			log.Printf("Error parsing size: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		size = s
	}

	if len(id) > 1 {
		code := id[0]
		id = id[1:]

		var emote emotes.Emote
		switch code {
		case 'b':
			e, found := store.FindBttvEmote(id)
			if found {
				emote = e
			} else {
				log.Printf("Requested BTTV emote %q but wasn't found\n", id)
				http.NotFound(w, r)
				return
			}
		case 'f':
			e, found := store.FindFfzEmote(id)
			if found {
				emote = e
			} else {
				log.Printf("Requested FFZ emote %q but wasn't found\n", id)
				http.NotFound(w, r)
				return
			}
		default:
			log.Printf("Requested emote of unknown type %q\n", code)
			http.NotFound(w, r)
			return
		}

		if cache == nil { // fallback to bttv cdn
			http.Redirect(w, r, emote.URL(size), http.StatusFound)
		} else { // use our own cache
			w.Header().Set("Content-Type", "image/png")
			err := cache.GetCachedOrDownload(emote, emotes.ImageSizeLarge, w)
			if err != nil {
				log.Printf("Error downloading emote: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
			}
		}
		return
	}

	log.Printf("Got unknown emote code %q\n", r.URL)
	http.NotFound(w, r)
}
