package main

import (
	"fmt"
	"os"

	bp3disk "github.com/moshenahmias/bp3/pkg/disk"
)

func main() {
	// Create or open files for store and page
	store, err := os.OpenFile("store.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Error opening store file:", err)
		return
	}
	defer store.Close()

	var pages []bp3disk.ReadWriteSeekSyncTruncater

	for i := 0; i < 10; i++ {
		page, err := os.OpenFile(fmt.Sprintf("page_%d.db", i), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("Error opening page file:", err)
			return
		}

		pages = append(pages, page)

		defer page.Close()
	}

	// Initialize a new B+ Tree instance
	tree, err := bp3disk.Initialize[int, string](
		store,
		pages[0],
		bp3disk.WithIndexPages(pages[1:]),
	)

	if err != nil {
		fmt.Println("Error initializing B+ Tree:", err)
		return
	}

	for i := 0; i < 1000000; i++ {
		// Insert some key-value pairs
		tree.Insert(i, fmt.Sprintf("value_%d", i))
	}

	// Flush the tree to the storage
	if err := bp3disk.Flush(tree); err != nil {
		fmt.Println("Error flushing B+ Tree:", err)
		return
	}

	// Load the B+ Tree from the storage
	loadedTree, err := bp3disk.Load[int, string](
		store,
		pages[0],
		bp3disk.WithIndexPages(pages[1:]),
	)

	if err != nil {
		fmt.Println("Error loading B+ Tree:", err)
		return
	}

	// Find a value in the loaded tree
	value, found := loadedTree.Find(100)
	if found {
		fmt.Println("Found value:", value)
	} else {
		fmt.Println("Value not found")
	}
}
