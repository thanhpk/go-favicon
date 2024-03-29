// Copyright (c) 2020 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT
// Created on 2020-11-09

package favicon

import (
	"io"
	urls "net/url"
	"path/filepath"
	"strings"

	gq "github.com/PuerkitoBio/goquery"
)

// entry point for URLs
func (p *parser) parseURL(url string) ([]*Icon, error) {
	u, err := urls.Parse(url)
	if err != nil {
		return nil, err
	}
	p.baseURL = u

	rc, err := p.find.fetchURL(url)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	doc, err := gq.NewDocumentFromReader(rc)
	if err != nil {
		return nil, err
	}
	return p.parse(doc), nil
}

// entry point for io.Reader
func (p *parser) parseReader(r io.Reader) ([]*Icon, error) {
	doc, err := gq.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	return p.parse(doc), nil
}

// main parser function
func (p *parser) parse(doc *gq.Document) []*Icon {
	var (
		icons       []*Icon
		manifestURL = p.absURL("/manifest.json")
	)

	// icons described in <link../> tags
	doc.Find("link").Each(func(i int, sel *gq.Selection) {
		rel, _ := sel.Attr("rel")
		rel = strings.ToLower(rel)

		switch rel {
		// all cases are handled the same way for now
		case "icon", "alternate icon", "shortcut icon":
			icons = append(icons, p.parseLink(sel)...)
		case "apple-touch-icon", "apple-touch-icon-precomposed":
			icons = append(icons, p.parseLink(sel)...)
		// site-specific browser apps (https://fluidapp.com/)
		case "fluid-icon":
			icons = append(icons, p.parseLink(sel)...)
		case "manifest":
			url, _ := sel.Attr("href")
			url = p.absURL(url)
			if url != "" {
				manifestURL = url
			}
		}
	})

	// retrieve and parse JSON manifest
	if !p.find.ignoreManifest {
		icons = append(icons, p.parseManifest(manifestURL)...)
	}
	// check for existence of URLs like /favicon.ico
	if !p.find.ignoreWellKnown {
		icons = append(icons, p.findWellKnownIcons()...)
	}

	return p.postProcessIcons(icons)
}

// extract icons defined in <link../> tags
func (p *parser) parseLink(sel *gq.Selection) []*Icon {
	var (
		href, _ = sel.Attr("href")
		typ, _  = sel.Attr("type")
		icons   []*Icon
		icon    = &Icon{}
	)

	if href = p.absURL(href); href == "" {
		return nil
	}

	icon.URL = href
	// icon.FileExt = fileExt(href)
	if typ != "" {
		icon.MimeType = typ
	}

	if len(icons) == 0 { // no sizes understood
		icons = append(icons, icon)
	}
	return icons
}

// extract file extension from a URL
func fileExt(url string) string {
	u, err := urls.Parse(url)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(filepath.Ext(u.Path), ".")
}
