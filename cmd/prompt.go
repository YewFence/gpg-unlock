package cmd

import (
	"bufio"
	"strings"
)

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func confirmYes(r *bufio.Reader) bool {
	line := readLine(r)
	return strings.EqualFold(line, "y") || strings.EqualFold(line, "yes")
}
