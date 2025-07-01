package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// ─── Entry point ────────────────────────────────────────────────────────────────
func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [directory]\n", filepath.Base(os.Args[0]))
	}
	flag.Parse()

	root := ".."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	info, err := os.Stat(root)
	if err != nil {
		log.Fatal(err)
	}
	if !info.IsDir() {
		log.Fatalf("%q is not a directory", root)
	}

	// Print the directory name itself and recurse into children.
	fmt.Println(info.Name())
	if err := walk(root, ""); err != nil {
		log.Fatal(err)
	}
}

// ─── Recursive walk (uses filepath.WalkDir under the hood) ──────────────────────
func walk(dir, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Sort so that directories appear first, then files, both alphabetically.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() == entries[j].IsDir() {
			return entries[i].Name() < entries[j].Name()
		}
		return entries[i].IsDir()
	})

	for i, entry := range entries {
		conn, nextPrefix := branch(i == len(entries)-1, prefix)
		fmt.Printf("%s%s%s\n", prefix, conn, entry.Name())

		if entry.IsDir() {
			// Recurse into sub-directory.
			if err := walk(filepath.Join(dir, entry.Name()), nextPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

// branch decides which connector symbols and next-level prefix to use.
func branch(isLast bool, prefix string) (connector, nextPrefix string) {
	if isLast {
		return "└── ", prefix + "    "
	}
	return "├── ", prefix + "│   "
}
