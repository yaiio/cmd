package main

import (
	"fmt"
	"github.com/yai/yai"
	"io/ioutil"
	"os"
	"path/filepath"
)

var cmdPackage = &Command{
	UsageLine: "package [import path]",
	Short:     "package a Yai application (e.g. for deployment)",
	Long: `
Package the Yai web application named by the given import path.
This allows it to be deployed and run on a machine that lacks a Go installation.

For example:

    yai package github.com/yai/samples/chat
`,
}

func init() {
	cmdPackage.Run = packageApp
}

func packageApp(args []string) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, cmdPackage.Long)
		return
	}

	appImportPath := args[0]
	yai.Init("", appImportPath, "")

	// Remove the archive if it already exists.
	destFile := filepath.Base(yai.BasePath) + ".tar.gz"
	os.Remove(destFile)

	// Collect stuff in a temp directory.
	tmpDir, err := ioutil.TempDir("", filepath.Base(yai.BasePath))
	panicOnError(err, "Failed to get temp dir")

	buildApp([]string{args[0], tmpDir})

	// Create the zip file.
	archiveName := mustTarGzDir(destFile, tmpDir)

	fmt.Println("Your archive is ready:", archiveName)
}
