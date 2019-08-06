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
	"regexp"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
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
		name, err := icccm.WmNameGet(X, clientid)

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

func main() {
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	windows, err := BuildProperties(X)
	if err != nil {
		panic(err)
	}
	r, err := regexp.Compile("Fire")
	if err != nil {
		panic(err)
	}
	for _, w := range windows {
		if r.FindString(w.Class) != "" {
			fmt.Println(w)
		}

	}
}
