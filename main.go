package main

import "github.com/myzkey/gh-repo-settings/cmd"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
