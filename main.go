package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unicode/utf8"
	"unsafe"

	"github.com/heliaxdev/namada-address-migrations/namada"
	"github.com/heliaxdev/namada-address-migrations/namada/addrconv"
)

type addressReplace struct {
	old []byte
	new []byte
}

var replacePool = sync.Pool{
	New: func() any {
		return make([]addressReplace, 0, 16)
	},
}

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
	case "i", "implicit-address":
		if len(os.Args) != 3 {
			return dieUsage()
		}
		return implicitAddress(os.Args[2])
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
			findReplace(path, ent)
		}()
		return nil
	})
}

func findReplace(path string, ent fs.DirEntry) {
	// 1 mib
	const sizeThreshold = 1 << 20

	if ent.Type()&fs.ModeType != 0 {
		return
	}

	info, err := ent.Info()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed read file info: %s: %s\n", path, err)
		return
	}
	if info.Size() >= sizeThreshold {
		return
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read file: %s: %s\n", path, err)
		return
	}

	if !utf8.Valid(buf) {
		return
	}

	occurrences := replacePool.Get().([]addressReplace)

	for _, match := range namada.AddressRegex.FindAll(buf, -1) {
		oldAddress := *(*string)(unsafe.Pointer(&match))
		newAddress, err := addrconv.ConvertAddress(oldAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to convert address: %s: %s\n", oldAddress, err)
			continue
		}
		occurrences = append(occurrences, addressReplace{
			old: bytes.Clone(match),
			new: []byte(newAddress),
		})
	}

	for i := 0; i < len(occurrences); i++ {
		occurrence := &occurrences[i]
		buf = bytes.ReplaceAll(buf, occurrence.old, occurrence.new)
	}

	occurrences = occurrences[:0]
	replacePool.Put(occurrences)

	err = os.WriteFile(path, buf, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to rewrite to file: %s: %s\n", path, err)
		return
	}
}

func convert(oldAddr string) error {
	newAddr, err := addrconv.ConvertAddress(oldAddr)
	if err != nil {
		return err
	}
	fmt.Println(newAddr)
	return nil
}

func implicitAddress(bech32PublicKey string) error {
	addr, err := addrconv.PublicKeyToImplicit(bech32PublicKey)
	if err != nil {
		return err
	}
	fmt.Println(addr)
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

	fmt.Fprintf(f, "\t# get an implicit address from a public key\n")
	fmt.Fprintf(f, "\ti|implicit-address <bech32-encoded-pk>\n")
	fmt.Fprintln(f)

	fmt.Fprintf(f, "\t# replace occurrences of old addresses with\n")
	fmt.Fprintf(f, "\t# the new address format in all files present\n")
	fmt.Fprintf(f, "\t# in the given path\n")
	fmt.Fprintf(f, "\tr|replace <path>\n")
	fmt.Fprintln(f)
}
