package main

import (
	"errors"
	"flag"
	"fmt"
	shared "github.com/JordanllHarper/distributed_file_backup/shared"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	sender := flag.String("senderAddr", "localhost:8080", "address to listen to")
	addr := flag.String("addr", "localhost:8081", "address to set the server to")
	dir := flag.String("dir", "", "where to store files")
	flag.Parse()
	go listenForFiles(*sender, *dir)

	fs := shared.NewFileStorage(*dir)

	http.HandleFunc("GET /v1/files/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		file, err := fs.Get(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(file)
	})

	http.ListenAndServe(*addr, nil)
}

func listenForFiles(sender, dir string) {
	senderAddr := fmt.Sprintf("http://%s%s", sender, "/v1/listen")
	res, err := http.DefaultClient.Get(senderAddr)
	if err != nil {
		log.Fatalln("Failed to send request to sender:", err)
	}
	defer res.Body.Close()

	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalln("Failed to get content type from the server")
	}
	boundary := params["boundary"]
	mr := multipart.NewReader(res.Body, boundary)
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			log.Println("Closed stream")
			return
		}

		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("Stream error:", err)
			return
		}

		filename := part.FileName()
		file, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
		if err != nil {
			log.Fatalln("Error creating file:", err)
		}
		written, err := io.Copy(file, part)
		if err != nil {
			log.Printf("Error writing to file %s: %s", filename, err)
			return
		}
		file.Close()
		log.Printf("Written %v to %v", written, filename)
	}
}
