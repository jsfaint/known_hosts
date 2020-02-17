/*
Package main implements an utility for manage the ssh known_hosts file.
A clone of https://github.com/markmcconachie/known_hosts write in Go.
Aimming to support Windows/Linux/macOS.
*/
package main

func main() {
	if !Exists() {
		return
	}

	h := ParseFromFile()
	Dump(h)
}
