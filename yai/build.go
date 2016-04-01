package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bitbucket.org/appsynth/go-platform/cmd/harness"
	"bitbucket.org/appsynth/go-platform/yai"
)

var cmdBuild = &Command{
	UsageLine: "build [import path] [target path]",
	Short:     "build a Yai application (e.g. for deployment)",
	Long: `
Build the Yai web application named by the given import path.
This allows it to be deployed and run on a machine that lacks a Go installation.

WARNING: The target path will be completely deleted, if it already exists!

For example:

    yai build github.com/yai/samples/chat /tmp/chat
`,
}

func init() {
	cmdBuild.Run = buildApp
}

func buildApp(args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "%s\n%s", cmdBuild.UsageLine, cmdBuild.Long)
		return
	}

	appImportPath, destPath := args[0], args[1]
	if !yai.Initialized {
		yai.Init("", appImportPath, "")
	}

	// First, verify that it is either already empty or looks like a previous
	// build (to avoid clobbering anything)
	if exists(destPath) && !empty(destPath) && !exists(path.Join(destPath, "run.sh")) {
		errorf("Abort: %s exists and does not look like a build directory.", destPath)
	}

	os.RemoveAll(destPath)
	os.MkdirAll(destPath, 0777)

	app, reverr := harness.Build()
	panicOnError(reverr, "Failed to build")

	// Included are:
	// - run scripts
	// - binary
	// - yai
	// - app

	// Yai and the app are in a directory structure mirroring import path
	srcPath := path.Join(destPath, "src")
	destBinaryPath := path.Join(destPath, filepath.Base(app.BinaryPath))
	tmpYaiPath := path.Join(srcPath, filepath.FromSlash(yai.YAI_IMPORT_PATH))
	mustCopyFile(destBinaryPath, app.BinaryPath)
	mustChmod(destBinaryPath, 0755)
	mustCopyDir(path.Join(tmpYaiPath, "conf"), path.Join(yai.YaiPath, "conf"), nil)
	mustCopyDir(path.Join(tmpYaiPath, "templates"), path.Join(yai.YaiPath, "templates"), nil)
	mustCopyDir(path.Join(srcPath, filepath.FromSlash(appImportPath)), yai.BasePath, nil)

	// Find all the modules used and copy them over.
	config := yai.Config.Raw()
	modulePaths := make(map[string]string) // import path => filesystem path
	for _, section := range config.Sections() {
		options, _ := config.SectionOptions(section)
		for _, key := range options {
			if !strings.HasPrefix(key, "module.") {
				continue
			}
			moduleImportPath, _ := config.String(section, key)
			if moduleImportPath == "" {
				continue
			}
			modulePath, err := yai.ResolveImportPath(moduleImportPath)
			if err != nil {
				yai.ERROR.Fatalln("Failed to load module %s: %s", key[len("module."):], err)
			}
			modulePaths[moduleImportPath] = modulePath
		}
	}
	for importPath, fsPath := range modulePaths {
		mustCopyDir(path.Join(srcPath, importPath), fsPath, nil)
	}

	tmplData, runShPath := map[string]interface{}{
		"BinName":    filepath.Base(app.BinaryPath),
		"ImportPath": appImportPath,
	}, path.Join(destPath, "run.sh")

	mustRenderTemplate(
		runShPath,
		filepath.Join(yai.YaiPath, "..", "cmd", "yai", "package_run.sh.template"),
		tmplData)

	mustChmod(runShPath, 0755)

	mustRenderTemplate(
		filepath.Join(destPath, "run.bat"),
		filepath.Join(yai.YaiPath, "..", "cmd", "yai", "package_run.bat.template"),
		tmplData)
}
