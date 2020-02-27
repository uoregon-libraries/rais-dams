package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
)

func pathify(elements ...string) string {
	var list = strings.Split(basePath, "/")
	list = append(list, elements...)
	return "/" + path.Join(list...)
}

func serve() {
	log.Printf("Base path is %q", pathify())
	log.Printf("Index path is %q", pathify("index"))

	// Never use the default ServeMux!
	var srv = http.NewServeMux()
	srv.HandleFunc(pathify(), redirect)
	srv.HandleFunc(pathify("index")+"/", renderIndex)
	srv.HandleFunc(pathify("asset")+"/", renderAsset)
	srv.HandleFunc(pathify("api")+"/", renderAPIForm)
	if basePath != "/" {
		srv.HandleFunc("/", notFound)
	}

	log.Println("Listening on port 12417")
	var err = http.ListenAndServe(":12417", srv)
	if err != nil {
		log.Printf("Error trying to serve http: %s", err)
	}
}

type indexData struct {
	Collection *collection
}

func redirect(w http.ResponseWriter, req *http.Request) {
	log.Printf("Request for %q", req.URL)
	if req.URL.Path == basePath {
		log.Printf("Redirecting %q to index", req.URL)
		http.Redirect(w, req, pathify("index"), 302)
	}
}

func getPathParts(req *http.Request) []string {
	var p = req.URL.RawPath
	if p == "" {
		p = req.URL.Path
	}

	var subPath = strings.Replace(p, pathify(basePath)+"/", "", 1)
	var parts = strings.Split(subPath, "/")

	// There's always an empty path part to strip off because of how pathify works
	return parts[1:]
}

func renderIndex(w http.ResponseWriter, req *http.Request) {
	// Strip "index" from the path
	var parts = getPathParts(req)[1:]

	var c = root
	if len(parts) > 0 && parts[0] != "" {
		var name string
		name, parts = parts[0], parts[1:]
		c = collections[name]

		if c == nil {
			log.Printf("Unable to find collection named %q (path %#v)", name, getPathParts(req))
			http.NotFound(w, req)
			return
		}
	}

	var data = indexData{Collection: c}
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
	return strings.Replace(p, pathify("asset")+"/", "", 1)
}

func findAsset(id string) *asset {
	for _, a2 := range jp2s {
		if a2.ID == id {
			return a2
		}
	}

	return nil
}

func renderAsset(w http.ResponseWriter, req *http.Request) {
	var id = findAssetID(req)
	if id == "" {
		http.Error(w, "invalid asset request", http.StatusBadRequest)
		return
	}

	var a = findAsset(id)
	if a == nil {
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

func notFound(w http.ResponseWriter, req *http.Request) {
	log.Printf("Path not handled: %q", req.URL)
	http.NotFound(w, req)
}
