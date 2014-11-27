package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

var pt = fmt.Printf

func main() {
	// take snapshots
	out := run("zfs", "list", "-o", "name", "-H")
	now := time.Now()
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		run("zfs", "snapshot", fmt.Sprintf("%s@autosnap-%04d-%02d-%02d-%02d-%02d-%02d",
			line, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))
	}
	// delete unchanged snapshot
	out = run("zfs", "list", "-t", "snapshot", "-o", "name,used", "-p", "-H")
	lines = strings.Split(out, "\n")
	var toDel []string
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if len(line) == 0 {
			break
		}
		parts := strings.SplitN(line, "\t", 2)
		if parts[1] == "0" {
			lastLine := strings.TrimSpace(lines[i-1])
			lastParts := strings.SplitN(lastLine, "\t", 2)
			out = run("zfs", "diff", lastParts[0], parts[0])
			if len(out) == 0 { // no change
				toDel = append(toDel, parts[0])
			}
		}
	}
	for _, name := range toDel {
		pt("delete snapshot %s\n", name)
		run("zfs", "destroy", name)
	}
}

func run(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		pt("%s\n", out)
		log.Fatal(err)
	}
	return string(out)
}
