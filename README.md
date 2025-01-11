# B+Tree Library in Go

## Description
This library provides an implementation of a B+Tree in Go. A B+Tree is a self-balancing tree data structure that maintains sorted data and allows for efficient insertion, deletion, and search operations. This implementation supports various order types and is generic, allowing it to work with any ordered key type and any value type.

## Usage
Here is a basic examples of how to use the library:

```go
package main

import (
    "fmt"
    "github.com/moshenahmias/bp3/pkg/bp3"
)

func main() {
    // Create a new B+Tree instance of order 3
    tree := bp3.New[int, string](bp3.WithOrder(3))

    // Insert elements
    tree.Insert(1, "one")
    tree.Insert(2, "two")
    tree.Insert(3, "three")
    tree.Insert(4, "four")
    tree.Insert(5, "five")

    // Search for elements
    value, found := tree.Find(2)
    if found {
        fmt.Println("Found:", value)
    } else {
        fmt.Println("Not Found")
    }

    // Delete elements
    deletedValue, deleted := tree.Delete(2)
    if deleted {
        fmt.Println("Deleted:", deletedValue)
    } else {
        fmt.Println("Not Found")
    }

    // Check if the tree is empty
    if tree.Empty() {
        fmt.Println("The tree is empty")
    } else {
        fmt.Println("The tree is not empty")
    }

    // Iterate elements
    for k, v := range tree.RangeClosed(2, 4) {
        fmt.Println("%d=%s", k, v)
    }
}
```

And with disk persistency support:

```go
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
```

## Installation
To install the library, use `go get`:
```sh
go get github.com/moshenahmias/bp3
```
