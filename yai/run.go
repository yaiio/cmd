package main

import (
	"github.com/yai/cmd/harness"
	"github.com/yai/yai"
	"strconv"
)

var cmdRun = &Command{
	UsageLine: "run [import path] [run mode] [port]",
	Short:     "run a Yai application",
	Long: `
Run the Yai web application named by the given import path.

For example, to run the chat room sample application:

    yai run github.com/yai/samples/chat dev

The run mode is used to select which set of app.conf configuration should
apply and may be used to determine logic in the application itself.

Run mode defaults to "dev".

You can set a port as an optional third parameter.  For example:

    yai run github.com/yai/samples/chat prod 8080`,
}

func init() {
	cmdRun.Run = runApp
}

func runApp(args []string) {
	if len(args) == 0 {
		errorf("No import path given.\nRun 'yai help run' for usage.\n")
	}

	// Determine the run mode.
	mode := "dev"
	if len(args) >= 2 {
		mode = args[1]
	}

	// Find and parse app.conf
	yai.Init(mode, args[0], "")
	yai.LoadMimeConfig()

	// Determine the override port, if any.
	port := yai.HttpPort
	if len(args) == 3 {
		var err error
		if port, err = strconv.Atoi(args[2]); err != nil {
			errorf("Failed to parse port as integer: %s", args[2])
		}
	}

	yai.INFO.Printf("Running %s (%s) in %s mode\n", yai.AppName, yai.ImportPath, mode)
	yai.TRACE.Println("Base path:", yai.BasePath)

	// If the app is run in "watched" mode, use the harness to run it.
	if yai.Config.BoolDefault("watch", true) && yai.Config.BoolDefault("watch.code", true) {
		yai.TRACE.Println("Running in watched mode.")
		yai.HttpPort = port
		harness.NewHarness().Run() // Never returns.
	}

	// Else, just build and run the app.
	yai.TRACE.Println("Running in live build mode.")
	app, err := harness.Build()
	if err != nil {
		errorf("Failed to build app: %s", err)
	}
	app.Port = port
	app.Cmd().Run()
}
