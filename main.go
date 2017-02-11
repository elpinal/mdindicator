package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
)

var httpAddr = flag.String("http", ":8080", "HTTP service address")

func usage() {
	log.Print("usage: mdindicator markdownfile")
	flag.PrintDefaults()
}

var header = []byte(`<!DOCTYPE html>
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

`)

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	if err := serve(flag.Arg(0)); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func serve(name string) error {
	p := &provider{name: name}
	if err := p.convert(); err != nil {
		return err
	}
	errch := make(chan error)
	go func() {
		err := p.watch()
		if err != nil {
			errch <- err
		}
	}()
	http.Handle("/", p)
	go func() {
		err := http.ListenAndServe(*httpAddr, nil)
		if err != nil {
			errch <- err
		}
	}()
	return <-errch
}

type provider struct {
	name string

	mu   sync.Mutex
	html []byte
}

func (p *provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(header)
	p.mu.Lock()
	w.Write(p.html)
	p.mu.Unlock()
}

func (p *provider) convert() error {
	input, err := ioutil.ReadFile(p.name)
	if err != nil {
		return err
	}
	html := blackfriday.MarkdownCommon(input)
	p.mu.Lock()
	p.html = html
	p.mu.Unlock()
	return nil
}

func (p *provider) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(filepath.Dir(p.name)); err != nil {
		return err
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == p.name && event.Op != fsnotify.Rename {
				if err := p.convert(); err != nil {
					return err
				}
			}
		case err := <-watcher.Errors:
			return errors.Wrap(err, "watcher")
		}
	}
}
