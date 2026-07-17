package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
		if err := cmd.RunServer(*port, *creds, *project, *maxFileSize); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "upload":
		fs := flag.NewFlagSet("upload", flag.ExitOnError)
		collection := fs.String("collection", "", "Firestore collection path")
		files := fs.String("files", "", "Comma-separated list of file paths")
		creds := fs.String("creds", "./serviceAccountKey.json", "Path to service account JSON")
		project := fs.String("project", "", "Firebase project ID")
		maxFileSize := fs.Int64("max-file-size", 1048576, "Per-file size limit in bytes")
		concurrency := fs.Int("concurrency", 10, "Max parallel Firestore writes")
		fs.Parse(os.Args[2:])
		var fileList []string
		if *files != "" {
			fileList = strings.Split(*files, ",")
		}
		if err := cmd.RunUpload(*collection, fileList, *creds, *project, *maxFileSize, *concurrency); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

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
