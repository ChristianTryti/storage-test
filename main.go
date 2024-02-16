package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

const (
	url           = "https://anakinprodsademo.blob.core.windows.net"
	containerName = "memes"
	blobName      = "meme.jpg"
	port          = ":8000"
)

func main() {
	http.HandleFunc("/", serveImage)

	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Create a Managed Identity credential
	credentials, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Println("Failed to create a managed identity credential:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client, err := azblob.NewClient(url, credentials, nil)
	if err != nil {
		log.Println("Unable to acquire blob client:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Download the blob
	downloadCtx, _ := context.WithTimeout(ctx, time.Second*5)
	get, err := client.DownloadStream(downloadCtx, containerName, blobName, nil)
	if err != nil {
		log.Println("Failed to create blob download stream:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	downloadedData := bytes.Buffer{}
	retryReader := get.NewRetryReader(ctx, &azblob.RetryReaderOptions{})
	_, err = downloadedData.ReadFrom(retryReader)
	if err != nil {
		log.Println("Unable to read from stream:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = retryReader.Close()
	if err != nil {
		log.Println("Unable to close stream:", err)
	}

	// Set the content type header
	w.Header().Set("Content-Type", "image/jpeg")

	// Write the blob content to the response writer
	_, err = io.Copy(w, &downloadedData)
	log.Println("Served memes")
	if err != nil {
		log.Println("Failed to write blob content to response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
