package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/russross/blackfriday"
)

var httpAddr = flag.String("http", ":8080", "http address")

func usage() {
	log.Print("usage: mdindicator markdownfile")
	flag.PrintDefaults()
}

var header = `
<!DOCTYPE html>
<meta charset="utf-8">
<title>mdindicator</title>
<style>body {margin: auto; width: 80%}</style>

`

func index(html []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, header + string(html))
	}
}

func main() {
	log.SetFlags(0)

	flag.Usage = usage
	flag.Parse()

	narg := flag.NArg()
	if narg != 1 {
		flag.Usage()
		os.Exit(1)
	}

	file := flag.Arg(0)
	input, err := ioutil.ReadFile(file)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	output := blackfriday.MarkdownCommon(input)

	http.HandleFunc("/", index(output))
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
