package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dosFormat  string = "\r\n"
	unixFormat string = "\n"
)

//GetFilePath returns the filepath of known_hosts
func GetFilePath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(h, ".ssh", "known_hosts"), nil
}

//Exists returns the file existence
func Exists() bool {
	name, err := GetFilePath()
	if err != nil {
		return false
	}

	if _, err := os.Stat(name); err == nil {
		return true
	} else {
		return os.IsExist(err)
	}
}

func getLinebreak() string {
	if runtime.GOOS == "windows" {
		return dosFormat
	} else {
		return unixFormat
	}
}

func stringToLine(input string) (lines []string) {
	tmp := strings.Split(input, unixFormat)

	for _, v := range tmp {
		lines = append(lines, strings.TrimSpace(v))
	}

	return lines
}

func ReadFile() []string {
	name, _ := GetFilePath()

	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil
	}

	return stringToLine(string(b))
}

func SaveFile(input []string) error {
	name, _ := GetFilePath()

	str := strings.Join(input, getLinebreak()) + getLinebreak()

	return ioutil.WriteFile(name, []byte(str), 0644)
}

//Search Host from list
func Search(input []string, pattern string) []string {
	var out []string

	for _, v := range input {
		if strings.Contains(v, pattern) {
			out = append(out, v)
		}
	}

	return out
}

//Delete Host from list
func Delete(input []string, pattern string) []string {
	var out []string

	for _, v := range input {
		if strings.Contains(v, pattern) {
			continue
		}

		if v == "" {
			continue
		}

		out = append(out, v)
	}

	return out
}
