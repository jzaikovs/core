package core

import (
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jzaikovs/core/loggy"
)

func init() {
	mime.AddExtensionType(".json", MIME_JSON)
}

// ServeFile this is just for development, file handling (CDN) better done by nginx or other
// TODO: there can be better alternative just to use http.FileServer
func ServeFile(out Output, path string) {
	if x, err := url.Parse(path); err == nil {
		path = x.Path
	}

	loggy.Trace.Println(path)

	f, err := os.OpenFile(filepath.Join("./www/", path), os.O_RDONLY, 0)
	if err != nil {
		out.Response(Response_Not_Found)
		out.Flush()
		return
	}
	defer f.Close()

	out.Response(Response_Ok)
	out.SetContentType(mime.TypeByExtension(filepath.Ext(f.Name())))

	// then write directly to response writer
	io.Copy(out, f)

	out.Flush() // flush response header
}
