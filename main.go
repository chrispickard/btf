// Example window-names fetches a list of all top-level client windows managed
// by the currently running window manager, and prints the name and size
// of each window.
//
// This example demonstrates how to use some aspects of the ewmh and icccm
// packages. It also shows how to use the xwindow package to find the
// geometry of a client window. In particular, finding the geometry is
// intelligent, as it includes the geometry of the decorations if they exist.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/chrispickard/btf/version"
	"github.com/mattn/go-shellwords"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	matches  = kingpin.Flag("match", "Class or Title to match").Short('m').Strings()
	excludes = kingpin.Flag("exclude", "Class or Title to exclude").Short('e').Strings()
	program  = kingpin.Arg("program", "program to launch if matching fails").String()
)

type Window struct {
	Class string
	Name  string
	Id    xproto.Window
}

func BuildProperties(X *xgbutil.XUtil) ([]*Window, error) {
	// Connect to the X server using the DISPLAY environment variable.
	// Get a list of all client ids.
	clientids, err := ewmh.ClientListGet(X)
	if err != nil {
		log.Fatal(err)
	}

	var windows []*Window

	// Iterate through each client, find its name and find its size.
	for _, clientid := range clientids {
		name, err := ewmh.WmNameGet(X, clientid)

		// If there was a problem getting _NET_WM_NAME or if its empty,
		// try the old-school version.
		if err != nil || len(name) == 0 {
			name, err = icccm.WmNameGet(X, clientid)

			// If we still can't find anything, give up.
			if err != nil || len(name) == 0 {
				return nil, err
			}
		}

		// If we still can't find anything, give up.
		if err != nil || len(name) == 0 {
			return nil, err
		}
		class, err := icccm.WmClassGet(X, clientid)
		if err != nil || len(name) == 0 {
			return nil, err
		}
		window := &Window{
			Class: class.Class,
			Name:  name,
			Id:    clientid,
		}
		windows = append(windows, window)
	}
	return windows, nil
}

// FocusWindow ...
func FocusWindow(X *xgbutil.XUtil, id xproto.Window) error {
	return ewmh.ActiveWindowReq(X, id)
}

// from https://www.calhoun.io/concatenating-and-building-strings-in-go/
func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

// largely from https://www.calhoun.io/concatenating-and-building-strings-in-go/
func buildRegex(excludes []string) (*regexp.Regexp, error) {
	var sb strings.Builder
	for _, str := range excludes {
		s := regexp.QuoteMeta(str)
		sb.WriteString(s)
	}
	return regexp.Compile(sb.String())
}

func main() {
	kingpin.Version(version.VERSION)
	kingpin.Parse()
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	r, err := buildRegex(*matches)
	if err != nil {
		log.Fatal(err)
	}
	excluder, err := buildRegex(*excludes)
	if err != nil {
		log.Fatal(err)
	}

	windows, err := BuildProperties(X)
	if err != nil {
		log.Fatal(err)
	}

	for _, w := range windows {
		if r.FindString(w.Name) != "" || r.FindString(w.Class) != "" {
			if excluder.FindString(w.Name) == "" || excluder.FindString(w.Class) == "" {
				FocusWindow(X, w.Id)
				os.Exit(0)
			}
		}

	}
	fmt.Println("not found, opening", *program)
	words, err := shellwords.Parse(*program)
	if err != nil {
		log.Fatal(err)
	}
	err = exec.Command(words[0], words[1:]...).Start()
	if err != nil {
		log.Fatal(err)
	}
}
