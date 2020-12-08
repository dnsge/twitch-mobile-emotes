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

func handleEmoticonRequest(w http.ResponseWriter, r *http.Request, store *emotes.EmoteStore) {
	// URL is in format of "/emoticons/v1/<id>/<size>
	parts := strings.Split(r.URL.Path, "/")
	id := parts[3]
	size, err := getSizeFromString(parts[4])
	if err != nil {
		log.Printf("Error parsing size: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(id) > 1 {
		code := id[0]
		id = id[1:]
		switch code {
		case 'b':
			e, found := store.FindBttvEmote(id)
			if found {
				http.Redirect(w, r, e.URL(size), http.StatusFound)
			} else {
				log.Printf("Requested BTTV emote %q but wasn't found\n", id)
				http.NotFound(w, r)
			}
			return
		case 'f':
			e, found := store.FindFfzEmote(id)
			if found {
				http.Redirect(w, r, e.URL(size), http.StatusFound)
			} else {
				log.Printf("Requested FFZ emote %q but wasn't found\n", id)
				http.NotFound(w, r)
			}
			return
		}
	}

	log.Printf("Got unknown emote code %q\n", r.URL)
	http.NotFound(w, r)
}
