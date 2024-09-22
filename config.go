package main

import (
	"io/fs"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"

	"plaque/discogs"
)

var (
	config *struct {
		Library struct {
			Root  string
			Queue string
		}
		Mpv struct {
			Args string
			// default: "$HOME/.local/state/mpv/watch_later"
			WatchLaterDir string `mapstructure:"watch_later_dir"`
		}
	}

	discogsEnabled bool
)

// Get path of p, relative to the binary. Returns error if p does not exist.
func getAbsPath(p string) (string, error) {
	prog, _ := os.Executable()

	abs := filepath.Join(filepath.Dir(prog), p)
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return abs, nil
}

func generateQueue(n int) []string {
	var all []string
	_ = filepath.WalkDir(config.Library.Root, func(path string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(config.Library.Root, path)
		if strings.Count(rel, "/") == 1 {
			all = append(all, rel)
		}
		return nil
	})

	items := make([]string, n)
	for i, r := range rand.Perm(len(all) - 1)[:n] {
		items[i] = all[r]
	}
	return items
}

func init() {
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

	// TODO: replace global config var with viper.GetViper()? callers would
	// have to call viper.GetViper().Get(s), which is dumb

	// if not New, the Viper from discogs will be inherited, and
	// discogs/config.toml will be preferentially loaded
	x := viper.New()

	x.AddConfigPath(".") // relative to this file
	prog, _ := os.Executable()
	x.AddConfigPath(filepath.Dir(prog)) // relative to wherever the binary is
	x.SetConfigName("config")
	x.SetConfigType("toml")

	x.SetDefault("mpv.args", "--mute=no --no-audio-display --pause=no --start=0%")
	x.SetDefault("mpv.watch_later_dir", os.ExpandEnv("$HOME/.local/state/mpv/watch_later"))

	// i am generally fine with keeping 2 separate config objects, so i
	// don't use MergeInConfig
	err := x.ReadInConfig()
	if err != nil {
		panic("No config found")
	}

	if err := x.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	for _, v := range []string{
		config.Library.Root,
		config.Library.Queue,
		// config.Mpv.Args,
		// config.Mpv.WatchLaterDir,
	} {
		if v == "" {
			log.Fatalln("empty fields found:\n", spew.Sdump(config))
		}
	}

	// relative path supplied; try to convert it to absolute path
	if _, err := os.Stat(config.Library.Queue); err != nil {
		abs, err := getAbsPath(config.Library.Queue)
		if err != nil { // absolute path does not exist; create it
			_ = os.WriteFile(
				filepath.Join(filepath.Dir(prog), abs),
				[]byte(strings.Join(generateQueue(1000), "\n")),
				0644,
			)
		}
		config.Library.Queue = abs
	}

	discogsEnabled = discogs.Config != nil &&
		discogs.Config.Username != "" &&
		discogs.Config.Key != ""
}
