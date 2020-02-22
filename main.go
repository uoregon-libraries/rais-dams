// Package main, along with the various *.go.html files, demonstrates a very
// simple (and ugly) asset server that reads all S3 assets in a given region
// and bucket, and serves up HTML pages which point to a IIIF server (RAIS, of
// course) for thumbnails and full-image views.
package main

import (
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

type asset struct {
	Filename string
	ID       string
	Title    string
}

var emptyAsset = asset{}

var jp2s []asset
var indexT, assetT, adminT *template.Template

func main() {
	readAssets()
	preptemplates()
	serve()
}

func readAssets() {
	var imgdir = "/var/local/images"
	var files, err = filepath.Glob(filepath.Join(imgdir, "*.jp2"))
	if err != nil {
		log.Fatalf("Unable to read images in %q: %s", imgdir, err)
	}

	for _, f := range files {
		var id = url.PathEscape(f)
		jp2s = append(jp2s, asset{Filename: f, ID: id, Title: "Untitled thing"})
	}
	log.Printf("Indexed %d assets", len(jp2s))
}

func preptemplates() {
	var _, err = os.Stat("./templates/layout.go.html")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Unable to load HTML layout: file does not exist.  Make sure you run the demo from the docker/s3demo folder.")
		} else {
			log.Printf("Error trying to open layout: %s", err)
		}
		os.Exit(1)
	}

	var root = template.New("layout")
	var layout = template.Must(root.ParseFiles("templates/layout.go.html"))
	indexT = template.Must(template.Must(layout.Clone()).ParseFiles("templates/index.go.html"))
	assetT = template.Must(template.Must(layout.Clone()).ParseFiles("templates/asset.go.html"))
	adminT = template.Must(template.Must(layout.Clone()).ParseFiles("templates/admin.go.html"))
}
