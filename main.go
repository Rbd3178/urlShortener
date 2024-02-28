package main

import (
	"encoding/json"
	"github.com/Rbd3178/redBlackTree/tree"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func shortestNext(s string) string {
	chars := []byte(s)

	for i := len(chars) - 1; i >= 0; i-- {
		if chars[i] < 'z' {
			chars[i]++
			return string(chars[:i+1])
		}
	}

	return s + "a"
}

type server struct {
	urls           tree.Tree[string, string]
	treeLock       sync.RWMutex
	pendingWriters sync.WaitGroup
	WGLock         sync.Mutex
}

func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.pendingWriters.Wait()
	s.treeLock.RLock()
	defer s.treeLock.RUnlock()

	alias := strings.TrimPrefix(r.URL.Path, "/go/")
	address, err := s.urls.At(alias)
	if err != nil {
		http.Error(w, "alias \""+alias+"\" doesn't exist", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, address, http.StatusSeeOther)
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		http.Error(w, "prefix is required", http.StatusBadRequest)
		return
	}

	nextPrefix := prefix[:len(prefix)-1] + string(prefix[len(prefix)-1]+1)

	s.pendingWriters.Wait()
	s.treeLock.RLock()
	defer s.treeLock.RUnlock()

	type aliasURLPair struct {
		Alias any
		URL   any
	}

	var elements []aliasURLPair
	for _, pair := range s.urls.Range(prefix, nextPrefix) {
		elements = append(elements, aliasURLPair{Alias: pair[0], URL: pair[1]})
	}

	jsonData, err := json.Marshal(elements)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (s *server) handleInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error when parsing form data", http.StatusBadRequest)
		return
	}

	alias := r.Form.Get("alias")
	address := r.Form.Get("url")

	if address == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if alias == "" {
        maxAlias, _, _ := s.urls.Max()
        alias = shortestNext(maxAlias)
    }

	parsed, err := url.Parse(address)
	if err != nil {
		http.Error(w, "wrong URL format", http.StatusBadRequest)
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		http.Error(w, "wrong URL format", http.StatusBadRequest)
		return
	}

	s.WGLock.Lock()
	s.pendingWriters.Add(1)
	s.WGLock.Unlock()
	defer func() {
		s.WGLock.Lock()
		s.pendingWriters.Done()
		s.WGLock.Unlock()
	}()

	s.treeLock.Lock()
	defer s.treeLock.Unlock()

	err = s.urls.Insert(alias, address)
	if err != nil {
		http.Error(w, "alias \""+alias+"\" is already taken", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("URL added successfully with alias \"" + alias + "\"\n"))
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alias := r.URL.Query().Get("alias")

	if alias == "" {
		http.Error(w, "alias is required", http.StatusBadRequest)
		return
	}

	s.WGLock.Lock()
	s.pendingWriters.Add(1)
	s.WGLock.Unlock()
	defer func() {
		s.WGLock.Lock()
		s.pendingWriters.Done()
		s.WGLock.Unlock()
	}()

	s.treeLock.Lock()
	defer s.treeLock.Unlock()

	err := s.urls.Delete(alias)
	if err != nil {
		http.Error(w, "alias \""+alias+"\" already doesn't exist", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	var urlShortener server
	urlShortener.urls = tree.Tree[string, string]{}
	http.HandleFunc("/go/", urlShortener.handleRedirect)
	http.HandleFunc("/search", urlShortener.handleSearch)
	http.HandleFunc("/add", urlShortener.handleInsert)
	http.HandleFunc("/delete", urlShortener.handleDelete)
	http.ListenAndServe(":8090", nil)
}
