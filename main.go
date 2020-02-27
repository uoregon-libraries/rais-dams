// Package main, along with the various *.go.html files, demonstrates a very
// simple (and ugly) asset server that reads all S3 assets in a given region
// and bucket, and serves up HTML pages which point to a IIIF server (RAIS, of
// course) for thumbnails and full-image views.
package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var imgdir = "/var/local/images"
var basePath string
var indexT, assetT, adminT *template.Template
var root *collection

func main() {
	basePath = "/"
	readAssets()
	preptemplates()
	serve()
}

func readAssets() {
	root = &collection{Title: "Root", ID: "", RelPath: ""}
	crawlAssets(root)
}

func crawlAssets(c *collection) error {
	var fullpath = filepath.Join(imgdir, c.RelPath)
	log.Printf("Crawling %q for collections / images", fullpath)

	var infos, err = ioutil.ReadDir(fullpath)
	if err != nil {
		return fmt.Errorf("reading dir %q: %w", c.RelPath, err)
	}

	for _, info := range infos {
		// Descend into dirs after creating a sub-collection
		if info.IsDir() {
			log.Printf("Adding collection %q as a child", info.Name())
			var subColl = c.addChild(info.Name())
			err = crawlAssets(subColl)
			if err != nil {
				return err
			}
		}

		// Don't bother with anything else - dirs and regular files only
		if !info.Mode().IsRegular() {
			continue
		}

		// Title files: our first piece of metadata
		if info.Name() == "title" {
			log.Println("Reading title for collection")
			var titleFile = filepath.Join(imgdir, c.RelPath, "title")
			var contents []byte
			contents, err = ioutil.ReadFile(titleFile)
			if err != nil {
				return fmt.Errorf("reading %q: %w", titleFile, err)
			}
			c.Title = strings.TrimSpace(string(contents))
		}

		if strings.HasSuffix(info.Name(), ".jp2") {
			log.Printf("Adding image asset %q to collection", info.Name())
			var asset = c.addAsset(info.Name())
			jp2s[asset.ID] = asset
		}
	}

	log.Printf("Indexed %d assets", len(c.JP2s))
	return nil
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
	root.Funcs(template.FuncMap{
		"BasePath": func() string {
			if basePath == "/" {
				return ""
			} else {
				return basePath
			}
		},
	})
	var layout = template.Must(root.ParseFiles("templates/layout.go.html"))
	indexT = template.Must(template.Must(layout.Clone()).ParseFiles("templates/index.go.html"))
	assetT = template.Must(template.Must(layout.Clone()).ParseFiles("templates/asset.go.html"))
	adminT = template.Must(template.Must(layout.Clone()).ParseFiles("templates/admin.go.html"))
}
