package watcher

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
)

type changeHandler interface {
	handleFileChange(string)
}

type processRunner struct {
	command []string
	cancel  context.CancelFunc
}

func newProcessRunner(command []string) *processRunner {
	path, err := exec.LookPath(command[0])
	if err != nil {
		log.Fatalf("failed to find %s", command[0])
	}
	command[0] = path
	return &processRunner{command: command}
}

func (runner *processRunner) handleFileChange(path string) {
	if runner.cancel != nil {
		runner.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner.cancel = cancel
	log.Printf("running: %v", strings.Join(runner.command, " "))
	cmd := exec.CommandContext(ctx, runner.command[0], runner.command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("%v failed to start with error: %v", runner.command, err)
		return
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("%v finished with error: %v", runner.command, err)
		return
	}
	log.Printf("%v finished successfully", runner.command)
}
