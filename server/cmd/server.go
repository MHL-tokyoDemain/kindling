package cmd

import "fmt"

func RunServer(port int, credsPath, projectID string, maxFileSize int64) error {
	fmt.Printf("kindling server — port=%d creds=%s project=%s maxFileSize=%d\n",
		port, credsPath, projectID, maxFileSize)
	fmt.Println("Server stub: not yet implemented")
	return nil
}

func RunUpload(collection string, files []string, credsPath, projectID string, maxFileSize int64, concurrency int) error {
	fmt.Printf("kindling upload — collection=%s files=%v creds=%s project=%s maxFileSize=%d concurrency=%d\n",
		collection, files, credsPath, projectID, maxFileSize, concurrency)
	fmt.Println("Upload stub: not yet implemented")
	return nil
}
