package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func serve() {
	http.HandleFunc("/", renderIndex)
	http.HandleFunc("/asset/", renderAsset)
	http.HandleFunc("/api/", renderAPIForm)

	var fileServer = http.FileServer(http.Dir("."))
	http.Handle("/osd/", fileServer)

	log.Println("Listening on port 12417")
	var err = http.ListenAndServe(":12417", nil)
	if err != nil {
		log.Printf("Error trying to serve http: %s", err)
	}
}

type indexData struct {
	Assets []asset
}

func renderIndex(w http.ResponseWriter, req *http.Request) {
	var path = req.URL.Path
	if path != "/" {
		http.NotFound(w, req)
		return
	}
	var data = indexData{Assets: jp2s}
	var err = indexT.Execute(w, data)
	if err != nil {
		log.Printf("Unable to serve index: %s", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func findAssetID(req *http.Request) string {
	var p = req.URL.RawPath
	if p == "" {
		p = req.URL.Path
	}
	var parts = strings.Split(p, "/")
	if len(parts) < 3 {
		log.Printf("Invalid path parts %#v", parts)
		return ""
	}

	return strings.Join(parts[2:], "/")
}

func findAsset(id string) asset {
	for _, a2 := range jp2s {
		if a2.ID == id {
			return a2
		}
	}

	return emptyAsset
}

func renderAsset(w http.ResponseWriter, req *http.Request) {
	var id = findAssetID(req)
	if id == "" {
		http.Error(w, "invalid asset request", http.StatusBadRequest)
		return
	}

	var a = findAsset(id)
	if a == emptyAsset {
		log.Printf("Invalid asset id %q", id)
		http.Error(w, fmt.Sprintf("Asset %q doesn't exist", id), http.StatusNotFound)
		return
	}

	var err = assetT.Execute(w, map[string]interface{}{"Asset": a})
	if err != nil {
		log.Printf("Unable to serve asset page: %s", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func renderAPIForm(w http.ResponseWriter, req *http.Request) {
	var err = adminT.Execute(w, nil)
	if err != nil {
		log.Printf("Unable to serve admin page: %s", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}
