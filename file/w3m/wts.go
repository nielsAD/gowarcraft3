// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3m

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// TriggerString recognition
var reWTS = regexp.MustCompile(`^STRING (\d+)$`)
var reTS = regexp.MustCompile(`^TRIGSTR_(\d+)$`)

// TriggerStrings from war3map.wts
func (m *Map) TriggerStrings() (map[int]string, error) {
	if m.ts == nil {
		wts, err := m.Archive.Open("war3map.wts")
		if err != nil {
			return nil, err
		}
		defer wts.Close()

		buf := bufio.NewReader(wts)

		if _, err := buf.Discard(1); err != nil && err != io.EOF {
			return nil, err
		}

		var ts = make(map[int]string)
		for {
			l, err := buf.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			match := reWTS.FindStringSubmatch(strings.TrimSpace(l))
			if len(match) < 2 {
				continue
			}

			id, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}

			for {
				p1, err := buf.ReadString('\n')
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(p1) == "{" {
					break
				} else if !strings.HasPrefix(p1, "//") {
					return nil, ErrBadFormat
				}
			}

			var sb strings.Builder
			for {
				l, err := buf.ReadString('\n')
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(l) == "}" {
					break
				}

				if sb.Len() > 0 {
					sb.WriteByte('\n')
				}
				sb.WriteString(strings.TrimRight(l, "\r\n"))
			}

			ts[id] = sb.String()
		}

		m.ts = ts
	}

	return m.ts, nil
}

// ExpandString expands trigger strings in s and returns the expanded string
func (m *Map) ExpandString(s string) (string, error) {
	ts, err := m.TriggerStrings()
	if err != nil {
		return "", err
	}

	match := reTS.FindStringSubmatch(s)
	if ts == nil || len(match) == 0 {
		return s, nil
	}

	id, err := strconv.Atoi(match[1])
	if err != nil {
		return "", err
	}

	return ts[id], nil
}
