package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const futureStateLabel = "Target, not implemented"

var futureHeading = regexp.MustCompile(`(?i)\b(future|planned|target)\b`)

func lintFutureState(path string, data []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s: read: %w", path, err)
	}

	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence || !isFutureHeading(trimmed) {
			continue
		}
		labelLine := nextNonBlank(lines, i+1)
		if labelLine < 0 || !strings.Contains(strings.ToLower(lines[labelLine]), strings.ToLower(futureStateLabel)) {
			return fmt.Errorf("%s:%d: future-state block must be followed by the label %q", path, i+1, futureStateLabel)
		}
	}
	return nil
}

func isFutureHeading(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	heading := strings.TrimLeft(line, "#")
	if heading == line || heading == "" || heading[0] != ' ' {
		return false
	}
	return futureHeading.MatchString(heading)
}

func nextNonBlank(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return i
		}
	}
	return -1
}
