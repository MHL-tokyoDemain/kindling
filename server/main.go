package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kindling/kindling/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "server":
		fs := flag.NewFlagSet("server", flag.ExitOnError)
		port := fs.Int("port", 9876, "Port to listen on")
		creds := fs.String("creds", "./serviceAccountKey.json", "Path to service account JSON")
		project := fs.String("project", "", "Firebase project ID")
		maxFileSize := fs.Int64("max-file-size", 1048576, "Per-file size limit in bytes")
		fs.Parse(os.Args[2:])
		cmd.RunServer(*port, *creds, *project, *maxFileSize)

	case "upload":
		fs := flag.NewFlagSet("upload", flag.ExitOnError)
		collection := fs.String("collection", "", "Firestore collection path (required)")
		creds := fs.String("creds", "./serviceAccountKey.json", "Path to service account JSON")
		project := fs.String("project", "", "Firebase project ID")
		maxFileSize := fs.Int64("max-file-size", 1048576, "Per-file size limit in bytes")
		concurrency := fs.Int("concurrency", 10, "Max parallel Firestore writes")
		fs.Parse(os.Args[2:])
		cmd.RunUpload(*collection, []string{}, *creds, *project, *maxFileSize, *concurrency)

	default:
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println("Usage: kindling <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  server    Start the HTTP server")
	fmt.Println("  upload    One-shot batch upload")
}
