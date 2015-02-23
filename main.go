package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	pt             = fmt.Printf
	snapshotFormat = regexp.MustCompile(`[a-zA-Z0-9]+@autosnap-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{2}-[0-9]{2}-[0-9]{2}`)
	snapshotFreqs  = [...]struct {
		Age  time.Duration
		Freq time.Duration
	}{
		{time.Hour * 24 * 32, time.Hour * 81},
		{time.Hour * 24 * 24, time.Hour * 27},
		{time.Hour * 24 * 16, time.Hour * 9},
		{time.Hour * 24 * 8, time.Hour * 3},
		{0, time.Minute},
	}
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
}

func process(name string) {
	// take snapshot
	now := time.Now()
	run("zfs", "snapshot", fmt.Sprintf("%s@autosnap-%04d-%02d-%02d-%02d-%02d-%02d",
		name, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))

	if rand.Intn(300) != 0 {
		return
	}

	// group and delete old snapshots
	out := run("zfs", "list", "-d", "1", "-t", "snapshot", "-H", "-o", "name", "-p", name)
	groups := make(map[time.Duration][]string)
	for _, snapshot := range strings.Split(out, "\n") {
		snapshot := strings.TrimSpace(snapshot)
		if len(snapshot) == 0 {
			continue
		}
		if !snapshotFormat.MatchString(snapshot) {
			continue
		}
		// parse snapshot time
		t, err := time.ParseInLocation(name+"@autosnap-2006-01-02-15-04-05", snapshot, time.Local)
		if err != nil {
			log.Fatal(err)
		}
		// group
		age := now.Sub(t)
		if age < 0 {
			log.Fatal("snapshot time parse error: now %v, parsed %v, age %v", now, t, age)
		}
		for _, freq := range snapshotFreqs {
			if age > freq.Age {
				slot := age / freq.Freq * freq.Freq
				groups[slot] = append(groups[slot], snapshot)
				break
			}
			continue
		}
	}
	for slot, snapshots := range groups {
		if len(snapshots) > 1 {
			pt("%v\n", slot)
			for _, snapshot := range snapshots[:len(snapshots)-1] {
				pt("delete %s\n", snapshot)
				run("zfs", "destroy", snapshot)
			}
			pt("\n")
		}
	}
}

func main() {
	t0 := time.Now()
	defer func() {
		pt("done in %v\n", time.Now().Sub(t0))
	}()
	// get pools
	out := run("zpool", "list", "-H", "-o", "name")
	for _, name := range strings.Split(out, "\n") {
		name := strings.TrimSpace(name)
		if len(name) == 0 {
			continue
		}
		pt("pool %s\n", name)
		process(name)
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
