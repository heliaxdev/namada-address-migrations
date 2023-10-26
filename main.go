package main

import (
	"fmt"
	"os"

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
	}

	return dieUsage()
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
}
