package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

func findOldestKnown(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lastLine := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lastLine = line
		}
	}

	type msg struct {
		ID string `json:"id"`
	}

	m := msg{}
	json.Unmarshal([]byte(lastLine), &m)

	return m.ID
}

func appendStruct(f *os.File, s interface{}) {
	encoded, _ := json.Marshal(s)
	f.Write(append(encoded, '\n'))
}
