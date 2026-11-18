package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	shared "github.com/JordanllHarper/distributed_file_backup/shared"
)

type httpError interface {
	error
	Status() int
}

type dependencies struct {
	shared.FileStore
}

func newDependencies(fs shared.FileStore) dependencies {
	return dependencies{
		FileStore: fs,
	}
}

func addEndpointsV1(mux *http.ServeMux, d dependencies) {
	updateCh := make(chan fileUpdate)
	mux.HandleFunc("PUT /v1/files/{filename}", handlePutFilesByFilename(d, updateCh))
	mux.HandleFunc("GET /v1/files/{filename}", handleGetFilesByFilename(d))
	mux.HandleFunc("GET /v1/listen", handleGetListen(updateCh))
}

type fileUpdate struct {
	filename string
	b        io.Reader
}

func handleGetListen(fileUpdates chan fileUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", mw.FormDataContentType())
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Connection", "keep-alive")

		clientGone := r.Context().Done()
		rc := http.NewResponseController(w)
		for {
			select {
			case <-clientGone:
				log.Println("Client disconnect")
				return
			case update := <-fileUpdates:
				fileWriter, err := mw.CreateFormFile("filename", update.filename)
				_, err = io.Copy(fileWriter, update.b)
				if err != nil {
					log.Println("Error writing", err)
					return
				}

				if err := rc.Flush(); err != nil {
					log.Println("Error flushing", err)
					return
				}
				log.Println("Flushed file to client")
			}
		}
	}
}

// TODO: authentication via API Key
func handlePutFilesByFilename(store shared.FileStorer, updateCh chan fileUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		var buf bytes.Buffer
		tee := io.TeeReader(r.Body, &buf)
		if err := store.Store(filename, tee); err != nil {
			handleErrHttp(w, err)
			return
		}

		updateCh <- fileUpdate{
			filename: filename,
			b:        &buf,
		}

		// TODO: perhaps we bundle all locations for all machines here
		w.Header().Set("Location", fmt.Sprintf("/v1/files/%s", filename))
		w.WriteHeader(http.StatusCreated)
	}
}

func handleGetFilesByFilename(store shared.FileGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		file, err := store.Get(filename)
		if err != nil {
			handleErrHttp(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(file)
	}
}

func handleErrHttp(w http.ResponseWriter, err error) {
	var httpErr httpError
	if ok := errors.As(err, &httpErr); ok {
		w.WriteHeader(httpErr.Status())
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	log.Println("An unexpected error occurred:", err)
}
