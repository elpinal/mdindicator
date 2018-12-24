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
    font-size: 0.66rem;
  }

  code {
    font-family: menlo, monospace;
    font-size: 0.9rem;
  }
</style>

`)

func main() {
	var httpAddr = flag.String("http", ":9999", "HTTP service address")
	var verbose = flag.Bool("verbose", false, "set verbose")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	log.Printf("http://localhost%s", *httpAddr)
	if err := serve(*httpAddr, filepath.Clean(flag.Arg(0)), *verbose); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func serve(addr, name string, verbose bool) error {
	if verbose {
		log.Printf("starting with %s", name)
	}
	p := &provider{name: name}
	if err := p.convert(); err != nil {
		return err
	}
	errch := make(chan error)
	go func() {
		err := p.watch(verbose)
		if err != nil {
			errch <- err
		}
	}()
	http.Handle("/", p)
	go func() {
		err := http.ListenAndServe(addr, nil)
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

func (p *provider) watch(verbose bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if verbose {
		log.Printf("watching %s", filepath.Dir(p.name))
	}
	if err := watcher.Add(filepath.Dir(p.name)); err != nil {
		return err
	}
	for {
		select {
		case event := <-watcher.Events:
			if filepath.Clean(event.Name) == p.name && event.Op != fsnotify.Rename {
				if verbose {
					log.Printf("caught event (%s): %v", p.name, event.Op)
				}
				if err := p.convert(); err != nil {
					return err
				}
			}
		case err := <-watcher.Errors:
			return errors.Wrap(err, "watcher")
		}
	}
}
