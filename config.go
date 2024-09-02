package main

import (
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
)

// https://github.com/Ragnaroek/run-slacker/blob/a7a9e3618a10ab7a6c099cbb4210ee0c9af1469a/run.go#L16

type Config struct {
	Discogs struct {
		Username   string `toml:"username"`
		Key        string `toml:"key"`
		MaxResults int    `toml:"max_results"`
	}

	// TODO: can toml infer?
	Library struct {
		Root  string `toml:"root"`
		Queue string `toml:"queue"`
		// Foo   string
	}

	Mpv struct {
		Args          string
		WatchLaterDir string
	}
}

// https://github.com/gami/simple_arch_example/blob/34fb11a31acc35fcb01a1e36c3ea1194bbe23074/config/config.go#L32
// TODO: https://github.com/spf13/viper#reading-config-files (esp for cfg creation)

var (
	config *Config
	once   sync.Once

	// TODO: can be declared in config

	// https://mpv.io/manual/master/#options-watch-later-dir
	WatchLaterDir = os.ExpandEnv("$HOME/.local/state/mpv/watch_later")
	mpvArgs       = strings.Fields("--mute=no --no-audio-display --pause=no --start=0%")
)

func init() {
	once.Do(func() {
		_, err := toml.DecodeFile("./config.toml", &config)
		if err != nil {
			panic(err)
		}

		for _, v := range []string{
			config.Discogs.Key,
			config.Discogs.Username,
			config.Library.Root,
			config.Library.Queue,
		} {
			if v == "" {
				log.Fatalln("empty fields found:\n", spew.Sdump(config))
			}
		}

		// TODO: check int values, set to sane defaults
		if config.Discogs.MaxResults == 0 {
			config.Discogs.MaxResults = 10
		}

		for _, p := range []string{
			config.Library.Root,
			config.Library.Queue,
		} {
			_, err := os.Stat(p)
			if err != nil { //|| !i.IsDir() {
				log.Fatalln("not a directory: ", p)
			}
		}

		// if _, err := os.ReadFile(config.Library.Queue); err != nil {
		// 	panic("no queue file")
		// }

		// https://github.com/Xe/x/blob/master/entropy/shannon.go
		l := len(config.Discogs.Key)

		charFreq := make(map[rune]float64)
		for _, i := range config.Discogs.Key {
			charFreq[i]++
		}

		var sum float64
		for _, c := range charFreq {
			f := c / float64(l)
			sum += f * math.Log2(f)
		}

		if int(math.Ceil(sum*-1))*l < 200 {
			panic("invalid key?")
		}

		// log.Println(config.Library.Foo)
	})
}
