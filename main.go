package main

// Main entry point for the app. Handles command-line options, starts the web
// listener and any import, etc

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
)

type URLList struct {
	Urls map[string][]string
	mu   sync.Mutex
}

var (
	httpPort    int
	stateFile   string
	botUsername string

	urlsUsed *URLList
	urlRegex *regexp.Regexp
)

func init() {
	urlRegex = regexp.MustCompile(`<(http([^\||>]*))`)
}

func main() {
	// Parse command-line options
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: ./slack-oldbot -port=8001\n")
		flag.PrintDefaults()
	}

	flag.IntVar(&httpPort, "port", 8001, "The HTTP port on which to listen")
	flag.StringVar(&stateFile, "stateFile", "state", "File to use for maintaining our markov chain state")
	flag.StringVar(&botUsername, "botUsername", "oldbot", "The name of the bot when it speaks")

	var importDir = flag.String("importDir", "", "The directory of a Slack export")
	var importChan = flag.String("importChan", "", "Optional channel to limit the import to")

	flag.Parse()

	if httpPort == 0 {
		flag.Usage()
		os.Exit(2)
	}

	urlsUsed = &URLList{
		Urls: make(map[string][]string),
	}

	// Import into the url list
	if *importDir != "" {
		err := StartImport(importDir, importChan)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Rebuild the url list from state
		err := urlsUsed.Load(stateFile)
		if err != nil {
			//log.Fatal(err)
			log.Printf("Could not load from '%s'. This may be expected.", stateFile)
		} else {
			log.Printf("Loaded previous state from '%s'.", stateFile)
		}
	}

	// Start the webserver
	StartServer(httpPort)
}

func extractUrls(s string) (urls []string) {
	matches := urlRegex.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if match[1] != "" {
			urls = append(urls, match[1])
		}
	}

	return
}

// Add a url
func (u *URLList) Add(url string, ts string) {
	u.mu.Lock()
	u.Urls[url] = append(u.Urls[url], ts)
	u.mu.Unlock()
}

// Get the uses of a url
func (u *URLList) Get(url string) []string {
	u.mu.Lock()
	uses := u.Urls[url]
	u.mu.Unlock()

	return uses
}

// Save the urllist to a file
func (u *URLList) Save(fileName string) error {
	// Open the file for writing
	fo, err := os.Create(fileName)
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// Create an encoder and dump to it
	u.mu.Lock()
	defer u.mu.Unlock()

	enc := gob.NewEncoder(fo)
	err = enc.Encode(u)
	if err != nil {
		return err
	}

	return nil
}

// Load the urllist from a file
func (u *URLList) Load(fileName string) error {
	// Open the file for reading
	fi, err := os.Open(fileName)
	if err != nil {
		return err
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	// Create a decoder and read from it
	u.mu.Lock()
	defer u.mu.Unlock()

	dec := gob.NewDecoder(fi)
	err = dec.Decode(u)
	if err != nil {
		return err
	}

	return nil
}
