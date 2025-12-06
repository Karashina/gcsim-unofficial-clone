package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func countQuotedInBlock(r io.Reader, startPattern, endPattern string) ([]string, error) {
	scanner := bufio.NewScanner(r)
	inBlock := false
	var block string
	startRe := regexp.MustCompile(startPattern)
	endRe := regexp.MustCompile(endPattern)
	for scanner.Scan() {
		line := scanner.Text()
		if !inBlock {
			if startRe.MatchString(line) {
				inBlock = true
				block += line + "\n"
			}
		} else {
			block += line + "\n"
			if endRe.MatchString(line) {
				break
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	// extract quoted strings from lines like: "value",
	lineRe := regexp.MustCompile(`^\s*"([^\"]*)"\s*,?\s*$`)
	out := []string{}
	for _, l := range regexp.MustCompile("\r?\n").Split(block, -1) {
		l = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(l, "")
		if l == "" {
			continue
		}
		m := lineRe.FindStringSubmatch(l)
		if len(m) > 1 {
			out = append(out, m[1])
		}
	}
	return out, nil
}

func countLinesInConstBlock(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	inConst := false
	count := 0
	startRe := regexp.MustCompile(`^\s*const\s*\(`)
	endRe := regexp.MustCompile(`^\s*\)`)
	commentRe := regexp.MustCompile(`^\s*//`)
	for scanner.Scan() {
		line := scanner.Text()
		if !inConst {
			if startRe.MatchString(line) {
				inConst = true
				continue
			}
		} else {
			if endRe.MatchString(line) {
				break
			}
			t := regexp.MustCompile(`^\s*$`).MatchString(line)
			c := commentRe.MatchString(line)
			if !t && !c {
				count++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func main() {
	repo := "."
	// weapon.go
	wpath := repo + "/pkg/core/keys/weapon.go"
	wf, err := os.Open(wpath)
	if err != nil {
		fmt.Println("error opening", wpath, err)
		return
	}
	defer wf.Close()
	wnames, err := countQuotedInBlock(wf, `var weaponNames = \[\]string\{`, `\}`)
	if err != nil {
		fmt.Println("error counting weaponNames", err)
		return
	}
	wcount := len(wnames)
	wconst, err := countLinesInConstBlock(wpath)
	if err != nil {
		fmt.Println("error counting weapon consts", err)
		return
	}
	fmt.Printf("weaponNames entries: %d\n", wcount)
	fmt.Printf("Weapon const lines: %d\n", wconst)
	fmt.Printf("diff: %d\n", wcount-wconst)

	// detailed check: list const names and compare
	wf2, err := os.Open(wpath)
	if err != nil {
		fmt.Println("error opening", wpath, err)
		return
	}
	defer wf2.Close()
	scanner := bufio.NewScanner(wf2)
	inConst := false
	constNames := []string{}
	startRe := regexp.MustCompile(`^\s*const\s*\(`)
	endRe := regexp.MustCompile(`^\s*\)`)
	identRe := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)`)
	for scanner.Scan() {
		line := scanner.Text()
		if !inConst {
			if startRe.MatchString(line) {
				inConst = true
				continue
			}
		} else {
			if endRe.MatchString(line) {
				break
			}
			t := regexp.MustCompile(`^\s*$`).MatchString(line)
			c := regexp.MustCompile(`^\s*//`).MatchString(line)
			if t || c {
				continue
			}
			m := identRe.FindStringSubmatch(line)
			if len(m) > 1 {
				constNames = append(constNames, m[1])
			}
		}
	}
	for i, name := range constNames {
		if i >= wcount {
			fmt.Printf("const index %d (%s) has no weaponNames entry\n", i, name)
			break
		}
		if wnames[i] == "" {
			fmt.Printf("weaponNames[%d] is empty (const %s)\n", i, name)
		}
	}

	// detect mismatches where normalized const name != weaponNames entry
	wordRe := regexp.MustCompile(`[A-Za-z]+`)
	normalize := func(s string) string {
		parts := wordRe.FindAllString(s, -1)
		for i := range parts {
			parts[i] = strings.ToLower(parts[i])
		}
		return strings.Join(parts, "")
	}
	for i, cname := range constNames {
		if i >= wcount {
			break
		}
		expected := normalize(cname)
		actual := strings.ToLower(wnames[i])
		if expected != actual {
			fmt.Printf("mismatch at index %d: const %s -> expected '%s', got '%s'\n", i, cname, expected, actual)
		}
	}

	// Print a generated weaponNames slice from constNames for easy patching
	fmt.Println("\n--- GENERATED weaponNames slice ---")
	fmt.Println("var weaponNames = []string{")
	for _, cname := range constNames {
		fmt.Printf("\t\"%s\",\n", normalize(cname))
	}
	fmt.Println("}")

	// print last 10 of weaponNames and constNames for inspection
	fmt.Println("--- last weaponNames ---")
	start := 0
	if wcount > 10 {
		start = wcount - 10
	}
	for i := start; i < wcount; i++ {
		fmt.Printf("%d: %s\n", i, wnames[i])
	}
	fmt.Println("--- last constNames ---")
	cstart := 0
	if len(constNames) > 10 {
		cstart = len(constNames) - 10
	}
	for i := cstart; i < len(constNames); i++ {
		fmt.Printf("%d: %s\n", i, constNames[i])
	}

	// charNames
	cpath := repo + "/pkg/core/keys/char.go"
	cf, err := os.Open(cpath)
	if err != nil {
		fmt.Println("error opening", cpath, err)
		return
	}
	defer cf.Close()
	cnames, err := countQuotedInBlock(cf, `var charNames = \[EndCharKeys\]string\{`, `\}`)
	if err != nil {
		fmt.Println("error counting charNames", err)
		return
	}
	ccount := len(cnames)
	// generated char consts
	gpath := repo + "/pkg/core/keys/keys_char_gen.go"
	gconst, err := countLinesInConstBlock(gpath)
	if err != nil {
		fmt.Println("error counting generated char consts", err)
		return
	}
	fmt.Printf("charNames entries (literals in char.go): %d\n", ccount)
	fmt.Printf("Char const lines (generated): %d\n", gconst)
	fmt.Printf("diff: %d\n", ccount-gconst)
}
