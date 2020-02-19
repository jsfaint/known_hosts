package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func ReadFile() []string {
	name, _ := GetFilePath()

	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil
	}

	var sep string

	if strings.Contains(string(b), "\r\n") {
		sep = "\r\n"
	} else {
		sep = "\n"
	}

	return strings.Split(string(b), sep)
}

func SaveFile(input []string) error {
	name, _ := GetFilePath()

	str := strings.Join(input, "\n")

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

		out = append(out, v)
	}

	return out
}
