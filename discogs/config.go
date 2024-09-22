// Config initialisation procedures

package discogs

import (
	"log"
	"math"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

var Config *struct {
	Username   string
	Key        string
	MaxResults int
}

func init() {
	_, here, _, _ := runtime.Caller(0)

	viper.AddConfigPath(".")                // cwd -- for test only
	viper.AddConfigPath(filepath.Dir(here)) // relative to this file

	// note: AddConfigPath can expand $HOME
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		panic("No config found")
	}

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	// // TODO: check int values, set to sane defaults
	// if Config.MaxResults == 0 {
	// 	Config.MaxResults = 10
	// }

	if Config.Key == "" {
		log.Println("no key")
		return
	}

	// https://github.com/Xe/x/blob/master/entropy/shannon.go
	l := len(Config.Key)

	charFreq := make(map[rune]float64)
	for _, i := range Config.Key {
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

	// fmt.Println("discogs config ok", viper.AllSettings())
}
