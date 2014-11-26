package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func main() {
	out, err := exec.Command("zfs", "list", "-o", "name", "-H").CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", out)
		log.Fatal(err)
	}
	now := time.Now()
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		out, err = exec.Command("zfs", "snapshot", fmt.Sprintf("%s@autosnap-%04d-%02d-%02d-%02d-%02d-%02d",
			line, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())).CombinedOutput()
		if err != nil {
			fmt.Printf("%s\n", out)
			log.Fatal(err)
		}
		//TODO auto delete
	}
}
