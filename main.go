package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"replace-addrs/namada"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return dieUsage()
	}

	switch os.Args[1] {
	case "h", "help", "u", "usage":
		usage(os.Stdout)
		return nil
	case "c", "convert":
		if len(os.Args) != 3 {
			return dieUsage()
		}
		return convert(os.Args[2])
	case "r", "replace":
		if len(os.Args) != 3 {
			return dieUsage()
		}
		return replace(os.Args[2])
		return nil
	}

	return dieUsage()
}

func replace(rootPath string) error {
	wg := &sync.WaitGroup{}
	semaphore := make(chan struct{}, runtime.NumCPU()<<1)
	runFindReplace(wg, semaphore, rootPath)
	wg.Wait()
	return nil
}

func runFindReplace(wg *sync.WaitGroup, sem chan struct{}, rootPath string) {
	filepath.WalkDir(rootPath, func(path string, ent fs.DirEntry, err error) error {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			findReplace(path)
		}()
		return nil
	})
}

func findReplace(path string) {
	fmt.Println(path)
}

func convert(oldAddr string) error {
	newAddr, err := namada.ConvertAddress(oldAddr)
	if err != nil {
		return err
	}
	fmt.Println(newAddr)
	return nil
}

func dieUsage() error {
	usage(os.Stderr)
	return fmt.Errorf("invalid cli arguments detected")
}

func usage(f *os.File) {
	fmt.Fprintf(f, "usage: %s <args>, where <args> =\n", os.Args[0])
	fmt.Fprintln(f)

	fmt.Fprintf(f, "\t# show this help message\n")
	fmt.Fprintf(f, "\thelp|usage|h|u\n")
	fmt.Fprintln(f)

	fmt.Fprintf(f, "\t# convert from the old to the new addr fmt,\n")
	fmt.Fprintf(f, "\t# printing the new address to stdout\n")
	fmt.Fprintf(f, "\tc|convert <old-address>\n")
	fmt.Fprintln(f)

	fmt.Fprintf(f, "\t# replace occurrences of old addresses with\n")
	fmt.Fprintf(f, "\t# the new address format in all files present\n")
	fmt.Fprintf(f, "\t# in the given path\n")
	fmt.Fprintf(f, "\tr|replace <path>\n")
	fmt.Fprintln(f)
}
