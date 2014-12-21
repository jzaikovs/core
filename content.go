package core

import (
	"io"
	"mime"
	"os"
	"path/filepath"
)

func init() {
	mime.AddExtensionType(".json", MIME_JSON)
}

//	NOTE: this is just for development, file handling (CDN) better done by nginx or other
func ServeFile(out Output, path string) {
	//Log.Info("serving file:", path)
	f, err := os.OpenFile("./www/"+path, os.O_RDONLY, 0)
	if err != nil {
		out.Response(Response_Not_Found)
		out.Flush()
		return
	}
	defer f.Close()

	out.Response(Response_Ok)
	out.SetContentType(mime.TypeByExtension(filepath.Ext(f.Name())))

	// then write directly to response writer
	io.Copy(out.ResponseWriter(), f)

	//out.Flush() // flush response header
}
