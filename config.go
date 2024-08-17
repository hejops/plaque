package main

import (
	"log"
	"math"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
)

// https://github.com/Ragnaroek/run-slacker/blob/a7a9e3618a10ab7a6c099cbb4210ee0c9af1469a/run.go#L16

type Config struct {
	Discogs Discogs
	Library Library
}

type Discogs struct {
	Username string `toml:"username"`
	Key      string `toml:"key"`
}

type Library struct {
	Root  string `toml:"root"`
	Queue string `toml:"queue"`
}

// https://github.com/gami/simple_arch_example/blob/34fb11a31acc35fcb01a1e36c3ea1194bbe23074/config/config.go#L32

var (
	config *Config
	once   sync.Once
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
				log.Fatal("empty fields found:\n", spew.Sdump(config))
				return
			}
		}

		for _, d := range []string{
			config.Library.Root,
			// config.Library.Queue,
		} {
			i, err := os.Stat(d)
			if err != nil {
				panic(err)
			}
			if !i.IsDir() {
				log.Fatal("not a directory:\n", d)
			}

		}

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
	})
}
