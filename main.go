package main

import (
	"fmt"
	"io"
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
	app      = kingpin.New("btf", "bring to front")
	list     = app.Flag("list", "list all properties").Short('l').Bool()
	matches  = app.Flag("match", "Class or Title to match").Short('m').Strings()
	excludes = app.Flag("exclude", "Class or Title to exclude").Short('e').Strings()
	program  = app.Arg("program", "program to launch if matching fails").String()
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
func PrintProperties(windows []*Window, w io.Writer) {
	for _, window := range windows {
		fmt.Fprintf(w, "%s %s %v\n", window.Class, window.Name, window.Id)
	}
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
	app.Version(version.VERSION)
	app.Parse(os.Args)
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	if !*list && *program == os.Args[0] {
		app.FatalUsage("either --list or <program> is required")
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

	if *list {
		PrintProperties(windows, os.Stdout)
		os.Exit(0)
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
