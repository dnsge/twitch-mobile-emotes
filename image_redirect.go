package tme

import (
	"fmt"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/session"
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
	case "3.0":
		return emotes.ImageSizeLarge, nil
	default:
		return -1, fmt.Errorf("Unknown image size %s\n", text)
	}
}

func handleEmoticonRequest(w http.ResponseWriter, r *http.Request, store *emotes.EmoteStore, cache *emotes.ImageFileCache) {
	// URL comes in format of "/emoticons/<version>/...
	if strings.HasPrefix(r.URL.Path, "/emoticons/v1/") {
		// URL is in format of "/emoticons/v1/<id>/<size>"
		v1Handler(w, r, store, cache)
	} else if strings.HasPrefix(r.URL.Path, "/emoticons/v2/") {
		// URL is in format of "/emoticons/v2/<id>/<format>/<theme_mode>/<size>"
		v2Handler(w, r, store, cache)
	} else {
		http.NotFound(w, r)
		return
	}
}

func v1Handler(w http.ResponseWriter, r *http.Request, store *emotes.EmoteStore, cache *emotes.ImageFileCache) {
	// URL is in format of "/emoticons/v1/<id>/<size>"
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 5 { // verify URL
		http.NotFound(w, r)
		return
	}

	id := parts[3]
	size, err := getSizeFromString(parts[4])
	if err != nil {
		log.Printf("Error parsing size: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commonHandler(w, r, store, cache, id, size, false)
}

func v2Handler(w http.ResponseWriter, r *http.Request, store *emotes.EmoteStore, cache *emotes.ImageFileCache) {
	// URL is in format of "/emoticons/v2/<id>/<format>/<theme_mode>/<size>"
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 7 { // verify URL
		http.NotFound(w, r)
		return
	}

	id := parts[3]
	size, err := getSizeFromString(parts[6])
	if err != nil {
		log.Printf("Error parsing size: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commonHandler(w, r, store, cache, id, size, true)
}

func commonHandler(w http.ResponseWriter, r *http.Request, store *emotes.EmoteStore, cache *emotes.ImageFileCache, id string, size emotes.ImageSize, gifSupport bool) {
	if len(id) < 2 {
		log.Printf("Got unknown emote code %q\n", r.URL)
		http.NotFound(w, r)
		return
	}

	code := id[0]

	if code == 'd' { // cache destroyer, discard characters
		id = id[session.CacheDestroyerSize+1:]
		code = id[0]
	}

	id = id[1:]

	isVirtual := code == 'v'
	var half emotes.VirtualHalf = -1
	if isVirtual {
		// At this point, 'id' is in the form of [l/r][emote_type][emote_id]
		if len(id) < 3 {
			log.Printf("Got unknown emote code %q\n", r.URL)
			http.NotFound(w, r)
			return
		}

		switch id[0] {
		case 'l':
			half = emotes.LeftHalf
		case 'r':
			half = emotes.RightHalf
		default:
			log.Printf("Requested Virtual emote with unknown side %q\n", id[0])
			http.NotFound(w, r)
			return
		}
		code = id[1]
		id = id[2:]
	}

	var emote emotes.Emote
	switch code {
	case 'b':
		e, found := store.GetBttvEmote(id)
		if found {
			emote = e
		} else {
			log.Printf("Requested BTTV emote %q but wasn't found\n", id)
			http.NotFound(w, r)
			return
		}
	case 'f':
		e, found := store.GetFfzEmote(id)
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

	if cache == nil || (gifSupport && emote.Type() == "gif") { // fallback to bttv cdn
		http.Redirect(w, r, emote.URL(size), http.StatusFound)
	} else { // use our own cache
		w.Header().Set("Content-Type", "image/png")

		var err error
		if isVirtual {
			err = cache.GetCachedOrDownloadHalf(emote, size, half, w)
		} else {
			err = cache.GetCachedOrDownload(emote, size, w)
		}

		if err != nil {
			log.Printf("Error downloading emote: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
