// playback and queue management
//
// for scrobbling, consider https://github.com/Feqzz/mpv-lastfm-scrobbler

package main

import (
	"io"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// https://github.com/picosh/pico/blob/4632c9cd3d7bc37c9c0c92bdc3dc8a64928237d8/tui/senpai.go#L10

// wrapper to call functions in a blocking manner
type rateCmd struct{}

// required methods for tea.ExecCommand

func (c *rateCmd) Run() error {
	// if resume, return early
	rateRelease(4319735)
	// TODO: remove from queue
	return nil
}
func (c *rateCmd) SetStderr(io.Writer) {}
func (c *rateCmd) SetStdin(io.Reader)  {}
func (c *rateCmd) SetStdout(io.Writer) {}

func play(dir string) tea.Cmd {
	mpv_args := strings.Split("--mute=no --no-audio-display --pause=no --start=0%", " ")
	mpv_args = append(mpv_args, dir)
	cmd := exec.Command("mpv", mpv_args...)

	return tea.Sequence(
		// if the altscreen is not used, new model is (inexplicably)
		// rendered before (above) mpv
		// tea.EnterAltScreen,

		tea.ExecProcess(cmd, nil),

		// tea.ExitAltScreen,
		// tea.ClearScreen,

		// // this func/Msg actually works (program will wait for 1 line
		// // of stdin, then proceed with put/post), but since it is not
		// // blocking, very bizarre behaviour will be observed (e.g. new
		// // selector is rendered on top of prompt)
		// func() tea.Msg {
		// 	rateRelease(4319735)
		// 	return nil
		// },

		tea.Exec(&rateCmd{}, nil),

		// tea.ExitAltScreen,
		// tea.ClearScreen,
	)
}
