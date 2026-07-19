package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kindling/kindling/cmd"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}

	switch args[1] {
	case "server":
		fs := flag.NewFlagSet("server", flag.ContinueOnError)
		port := fs.Int("port", envInt("KINDLING_PORT", 9876), "Port to listen on")
		creds := fs.String("creds", envString("GOOGLE_APPLICATION_CREDENTIALS", "./serviceAccountKey.json"), "Path to service account JSON")
		project := fs.String("project", envString("FIREBASE_PROJECT_ID", ""), "Firebase project ID")
		maxFileSize := fs.Int64("max-file-size", envInt64("KINDLING_MAX_FILE_SIZE", 1048576), "Per-file size limit in bytes")
		if err := fs.Parse(args[2:]); err != nil {
			return 2
		}
		if err := cmd.RunServer(*port, *creds, *project, *maxFileSize); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		return 0

	case "upload":
		fs := flag.NewFlagSet("upload", flag.ContinueOnError)
		collection := fs.String("collection", "", "Firestore collection path")
		files := fs.String("files", "", "Comma-separated list of file paths")
		creds := fs.String("creds", envString("GOOGLE_APPLICATION_CREDENTIALS", "./serviceAccountKey.json"), "Path to service account JSON")
		project := fs.String("project", envString("FIREBASE_PROJECT_ID", ""), "Firebase project ID")
		maxFileSize := fs.Int64("max-file-size", envInt64("KINDLING_MAX_FILE_SIZE", 1048576), "Per-file size limit in bytes")
		concurrency := fs.Int("concurrency", 10, "Max parallel Firestore writes")
		if err := fs.Parse(args[2:]); err != nil {
			return 2
		}

		var fileList []string
		if *files != "" {
			fileList = strings.Split(*files, ",")
		}

		exitCode, err := cmd.RunUpload(*collection, fileList, *creds, *project, *maxFileSize, *concurrency)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return exitCode

	default:
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Println("Usage: kindling <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  server    Start the HTTP server")
	fmt.Println("  upload    One-shot batch upload")
}

func envString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func envInt64(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}
