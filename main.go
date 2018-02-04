package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/mightyguava/monty/livereload"
	"github.com/mightyguava/monty/subproc"
	"github.com/rjeczalik/notify"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("monty: ")
	log.SetOutput(os.Stderr)

	urlFlag := flag.String("url", "", "a URL to open in the browser and to live reload")
	flag.Parse()
	fmt.Println(*urlFlag)

	cmd := flag.Args()
	if len(cmd) == 0 && *urlFlag == "" {
		fmt.Println()
		os.Exit(1)
	}

	w, err := CreateWatcher()
	if err != nil {
		log.Fatal("error starting file watcher: ", err)
	}

	var r *subproc.Runner
	if len(cmd) > 0 {
		var args []string
		if len(cmd) > 1 {
			args = cmd[1:]
		}
		cmdString := strings.Join(cmd, " ")
		executable := cmd[0]
		r = subproc.NewRunner(exec.Command(executable, args...), cmdString)
		if err = r.Start(); err != nil {
			log.Fatal("error starting command: ", err.Error())
		}
	}

	var chrome *livereload.Chrome
	if *urlFlag != "" {
		url, err := url.Parse(*urlFlag)
		if err != nil {
			log.Fatal("invalid url: ", *urlFlag)
		}
		if url.Scheme == "" {
			url.Scheme = "http"
		}
		if chrome, err = livereload.NewChrome(url.String()); err != nil {
			log.Fatal("could not connect to chrome: ", err)
		}
		log.Println("opening Chrome to: ", url.String())
		if err = chrome.Open(); err != nil {
			log.Fatal("could not open url: ", url.String())
		}
	}

	reloader := NewReloader(r, chrome, w)
	if err = reloader.WatchAndRun(); err != nil {
		log.Fatal(err)
	}
}

// CreateWatcher creates and returns a fs watcher for the current working directory
func CreateWatcher() (chan notify.EventInfo, error) {
	c := make(chan notify.EventInfo, 100)
	if err := notify.Watch("./...", c, notify.All); err != nil {
		return nil, err
	}
	return c, nil
}
