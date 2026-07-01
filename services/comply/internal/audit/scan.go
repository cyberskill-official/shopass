package audit

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Finding struct {
	File string
	Line int
	Rule string
	Hint string
}

// Scan quet cay ma, bo qua dong co // audit:allow.
func Scan(root string) ([]Finding, error) {
	var out []Finding

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".sql" && ext != ".yaml" && ext != ".yml" && ext != ".sh" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "// audit:allow") || strings.Contains(line, "-- audit:allow") || strings.Contains(line, "# audit:allow") {
				lineNum++
				continue
			}

			for _, rule := range bannedRules {
				if rule.Pattern.MatchString(line) {
					out = append(out, Finding{
						File: path,
						Line: lineNum,
						Rule: rule.Name,
						Hint: rule.Hint,
					})
				}
			}
			lineNum++
		}
		return scanner.Err()
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].File == out[j].File {
			if out[i].Line == out[j].Line {
				return out[i].Rule < out[j].Rule
			}
			return out[i].Line < out[j].Line
		}
		return out[i].File < out[j].File
	})

	return out, nil
}
