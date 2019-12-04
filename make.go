package main

import (
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "build.go")
	cmd.Dir = "support/tasks"
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
