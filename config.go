// Config initialisation procedures

package main

import (
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

// https://github.com/Ragnaroek/run-slacker/blob/a7a9e3618a10ab7a6c099cbb4210ee0c9af1469a/run.go#L16

// TODO: switch from toml to viper https://github.com/spf13/viper#reading-config-files (esp for defaults)

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

var (
	config *Config
	// once   sync.Once

	// TODO: can be declared in config (with defaults)

	// https://mpv.io/manual/master/#options-watch-later-dir
	WatchLaterDir = os.ExpandEnv("$HOME/.local/state/mpv/watch_later")
	MpvArgs       = strings.Fields("--mute=no --no-audio-display --pause=no --start=0%")
)

func getAbsPath(rel string) string {
	abs := filepath.Join(filepath.Base(os.Args[0]), rel)
	if _, err := os.Stat(abs); err != nil {
		panic(err)
	}
	return abs
}

func init() {
	// TODO: if mpv running, exit

	// `init` is reserved keyword -- https://go.dev/ref/spec#Package_initialization
	//
	// Once.Do is guaranteed to run only once. this not terribly important
	// for the program (since our init process is quick, and doesn't
	// require any concurrency), but it makes sense within a getter func
	// (e.g. for a db connection). in our case, callers just access the
	// global var directly, so init is good enough
	//
	// https://medium.easyread.co/just-call-your-code-only-once-256f69ed39a8?gi=3f3afe51e2a4
	// https://github.com/gami/simple_arch_example/blob/34fb11a31acc35fcb01a1e36c3ea1194bbe23074/config/config.go#L32

	// note: both viper and toml suffer from relpath issue; specifically,
	// tests will be run in /tmp, where config.toml cannot be found.
	// however, viper makes it easy to check multiple paths

	// fmt.Println("config init")

	viper.AddConfigPath(".") // relative to this file
	prog, _ := os.Executable()
	viper.AddConfigPath(filepath.Dir(prog)) // relative to wherever the binary is
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		panic("No config found")
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	for _, v := range []string{
		// config.Discogs.Key,
		// config.Discogs.Username,
		config.Library.Root,
		config.Library.Queue,
	} {
		if v == "" {
			log.Fatalln("empty fields found:\n", spew.Sdump(config))
		}
	}

	if _, err := os.Stat(config.Library.Queue); err != nil {
		config.Library.Queue = getAbsPath(config.Library.Queue)
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

	// if _, err := os.ReadFile(c.Library.Queue); err != nil {
	// 	panic("no queue file")
	// }

	if config.Discogs.Key == "" {
		return
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

	// log.Println(c.Library.Foo)
}

// func init() {
// 	// config.Once.Do(config.init) // config is still nil
// 	once.Do(config.init) // method runs, but seems to have no effect
// }
