package main

import (
	"bytes"
	"errors"
	"fmt"
	"go.guoyk.net/common"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func exit(err *error) {
	if *err != nil {
		log.Println((*err).Error())
		os.Exit(1)
	}
}

type M map[string]interface{}

var linePattern = regexp.MustCompile(`(.+),(.+)%,(.+)%`)

var alert string

func main() {
	var err error
	defer exit(&err)

	out := &bytes.Buffer{}
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "{{.Name}},{{.CPUPerc}},{{.MemPerc}}")
	cmd.Stdout = out
	if err = cmd.Start(); err != nil {
		return
	}
	if err = cmd.Wait(); err != nil {
		return
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		fields := linePattern.FindStringSubmatch(line)
		if len(fields) != 4 {
			err = errors.New("bad line")
			return
		}
		name, cpuPerc, memPerc := strings.TrimSpace(fields[1]), strings.TrimSpace(fields[2]), strings.TrimSpace(fields[3])
		var cpu, mem float64
		cpu, _ = strconv.ParseFloat(cpuPerc, 10)
		mem, _ = strconv.ParseFloat(memPerc, 10)
		_ = cpu
		if mem > 80 {
			alert += fmt.Sprintf("%s, CPU = %.2f%%, MEM = %.2f%%\n", name, cpu, mem)
		}
	}

	if len(alert) > 0 {
		_ = common.PostJSON(os.Getenv("ALERT_URL"), M{"msgtype": "text", "text": M{"content": alert}}, nil)
	}
}
