package common

import (
	"fmt"
	"path/filepath"
	"strings"
)

var Emotes map[string]string

func EmoteToHtml(file, title string) string {
	return fmt.Sprintf(`<img src="/emotes/%s" height="28px" title="%s" />`, file, title)
}

func ParseEmotesArray(words []string) []string {
	newWords := []string{}
	for _, word := range words {
		// make :emote: and [emote] valid for replacement.
		wordTrimmed := strings.Trim(word, ":[]")

		found := false
		for key, val := range Emotes {
			if key == wordTrimmed {
				newWords = append(newWords, EmoteToHtml(val, key))
				found = true
			}
		}
		if !found {
			newWords = append(newWords, word)
		}
	}
	return newWords
}

func ParseEmotes(msg string) string {
	words := ParseEmotesArray(strings.Split(msg, " "))
	return strings.Join(words, " ")
}

func LoadEmotes() (int, error) {
	newEmotes := map[string]string{}

	emotePNGs, err := filepath.Glob("./static/emotes/*.png")
	if err != nil {
		return 0, fmt.Errorf("unable to glob emote directory: %s\n", err)
	}

	emoteGIFs, err := filepath.Glob("./static/emotes/*.gif")
	if err != nil {
		return 0, fmt.Errorf("unable to glob emote directory: %s\n", err)
	}
	globbed_files := []string(emotePNGs)
	globbed_files = append(globbed_files, emoteGIFs...)

	LogInfoln("Loading emotes...")
	emInfo := []string{}
	for _, file := range globbed_files {
		file = filepath.Base(file)
		key := file[0 : len(file)-4]
		newEmotes[key] = file
		emInfo = append(emInfo, key)
	}
	Emotes = newEmotes
	LogInfoln(strings.Join(emInfo, " "))
	return len(Emotes), nil
}
