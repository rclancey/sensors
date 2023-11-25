package sonos

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sync"

	"github.com/rclancey/httpserver/v2"
)

type FileServer struct {
	cfg *httpserver.NetworkConfig
	server *http.Server
	files map[string]string
	mutex *sync.Mutex
}

func NewFileServer(cfg *httpserver.NetworkConfig) (*FileServer, error) {
	port, ok := findFreePort(12000, 12200)
	if !ok {
		return nil, errors.New("no port available")
	}
	fs := &FileServer{
		cfg: cfg,
		files: map[string]string{},
		mutex: &sync.Mutex{},
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		fn, ok := fs.GetFileName(id)
		if !ok {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
			return
		}
		http.ServeFile(w, r, fn)
	}
	fs.server = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}
	go func() {
		fs.server.ListenAndServe()
		fs.server = nil
	}()
	return fs, nil
}

func (fs *FileServer) ServeFile(fileName string) string {
	id := fs.FileID(fileName)
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.files[id] = fileName
	return id
}

func (fs *FileServer) GetFileName(id string) (string, bool) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fn, ok := fs.files[id]
	return fn, ok
}

func (fs *FileServer) FileID(fileName string) string {
	hash := sha256.Sum256([]byte(fileName))
	return hex.EncodeToString(hash[:16]) + filepath.Ext(fileName)
}

func (fs *FileServer) FileURL(fileName string) string {
	ip := fs.cfg.GetIP()
	u := &url.URL{
		Scheme: "http",
		Host: ip.String() + fs.server.Addr,
		Path: "/" + fs.FileID(fileName),
	}
	return u.String()
}

func (fs *FileServer) FileForURL(uri string) (string, bool) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", false
	}
	id := path.Base(u.Path)
	return fs.GetFileName(id)
}
