package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

var collections = make(map[string]*collection)
var jp2s = make(map[string]*asset)

type asset struct {
	Parent   *collection
	Filename string
	ID       string
	Title    string
}

type collection struct {
	ID          string
	RelPath     string
	Parent      *collection
	Collections []*collection
	JP2s        []*asset
	Title       string
}

func (c *collection) addChild(name string) *collection {
	var child = &collection{RelPath: filepath.Join(c.RelPath, name), Parent: c}
	c.Collections = append(c.Collections, child)
	child.ID = url.PathEscape(child.RelPath)
	collections[child.ID] = child
	return child
}

func (c *collection) addAsset(name string) *asset {
	var asset = &asset{Parent: c, Filename: name, Title: "Untitled Asset"}
	var fullPath = filepath.Join(imgdir, c.RelPath, name)
	asset.ID = url.PathEscape(strings.Replace(fullPath, imgdir+"/", "", 1))
	c.JP2s = append(c.JP2s, asset)
	return asset
}
