package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/net/html"
)

type ImageScraper struct {
	AllowHost  []string
	visited    map[string]bool
	httpClient *http.Client
	dir        string
}

func New(dir string) *ImageScraper {
	return &ImageScraper{
		visited:    map[string]bool{},
		httpClient: http.DefaultClient,
		dir:        dir,
	}
}

func (s *ImageScraper) isAllowed(u *url.URL) bool {
	hp := net.JoinHostPort(u.Hostname(), u.Port())
	for _, h := range s.AllowHost {
		if h == hp {
			return true
		}
	}
	return false
}

func (s *ImageScraper) Visit(u *url.URL) error {

	urlStr := u.String()
	if s.visited[urlStr] {
		return nil
	}

	if !s.isAllowed(u) {
		return nil
	}

	s.visited[urlStr] = true

	fmt.Println("Visit", urlStr)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if err := s.parse(u, bytes.NewReader(body)); err != nil {
		return err
	}

	return nil
}

func (s *ImageScraper) parse(baseURL *url.URL, r io.Reader) error {
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}

	if err := s.traverse(baseURL, doc); err != nil {
		return err
	}

	return nil
}

func (s *ImageScraper) attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func (s *ImageScraper) traverse(baseURL *url.URL, n *html.Node) error {

	switch {
	case n.Type == html.ElementNode && n.Data == "a":
		if urlStr := s.attr(n, "href"); urlStr != "" {
			absURL, err := s.absoluteURL(baseURL, urlStr)
			if err != nil {
				return err
			}

			if err := s.Visit(absURL); err != nil {
				return err
			}
		}
	case n.Type == html.ElementNode && n.Data == "img":
		if src := s.attr(n, "src"); src != "" {
			absURL, err := s.absoluteURL(baseURL, src)
			if err != nil {
				return err
			}

			if err := s.downloadImage(absURL); err != nil {
				return err
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := s.traverse(baseURL, c); err != nil {
			return err
		}
	}

	return nil
}

func (s *ImageScraper) absoluteURL(baseURL *url.URL, ref string) (*url.URL, error) {
	refURL, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	return baseURL.ResolveReference(refURL), nil
}

func (s *ImageScraper) downloadImage(srcURL *url.URL) error {
	req, err := http.NewRequest(http.MethodGet, srcURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// copy
	u := *srcURL
	u.RawQuery = ""
	path := filepath.Join(s.dir, path.Base(u.String()))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}
