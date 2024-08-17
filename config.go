package main

import (
	"sync"

	"github.com/BurntSushi/toml"
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
	Root string `toml:"root"`
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
	})
}
