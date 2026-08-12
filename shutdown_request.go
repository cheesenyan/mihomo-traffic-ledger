package main

const shutdownCommand = "--shutdown"

func isShutdownCommand(args []string) bool {
	return len(args) == 2 && args[1] == shutdownCommand
}
