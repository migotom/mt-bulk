package mtbulkrestapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/migotom/mt-bulk/internal/entities"

	"github.com/migotom/mt-bulk/internal/clients"
)

// StripFileIndexes is simple middleware to filter file indexes.
func (mtbulk *MTbulkRESTGateway) StripFileIndexes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// FileUpload uploads a file into RootDirectory.
func (mtbulk *MTbulkRESTGateway) FileUpload(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(32 << 20) // 32Mb
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileName, err := clients.SecurePathJoin(mtbulk.Config.RootDirectory, header.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fh, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0744)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer fh.Close()

		_, err = io.Copy(fh, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result := entities.Result{DownloadURLs: []string{fileName}}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
