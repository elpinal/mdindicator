package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/russross/blackfriday"
)

var httpAddr = flag.String("http", ":8080", "HTTP service address")

func usage() {
	log.Print("usage: mdindicator markdownfile")
	flag.PrintDefaults()
}

var header = `
<!DOCTYPE html>
<meta charset="utf-8">
<title>mdindicator</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, sans-serif;
    font-size: 1rem;
    line-height: 1.5;
    margin: auto;
    width: 80%;
  }

  pre {
    background-color: #f7f7f7;
    border-radius: 4px;
    padding: 1rem;
  }

  pre > code {
    font-family: menlo, monospace;
    font-size: 0.66rem;
  }
</style>

`

var html string

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, header+html)
}

func convert(file string) error {
	input, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	output := blackfriday.MarkdownCommon(input)
	html = string(output)

	return nil
}

func main() {
	log.SetFlags(0)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	file := flag.Arg(0)
	if err := convert(file); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	go watch(file)

	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

func watch(file string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Name == file && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
					if err := convert(file); err != nil {
						log.Print(err)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("watcher: %v", err)
			}
		}
	}()

	if err := watcher.Add(filepath.Dir(file)); err != nil {
		log.Fatal(err)
	}
	<-done
	log.Print("after done")
}
