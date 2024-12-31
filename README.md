# B+Tree Library in Go

## Description
This library provides an implementation of a B+Tree in Go. A B+Tree is a self-balancing tree data structure that maintains sorted data and allows for efficient insertion, deletion, and search operations. This implementation supports various order types and is generic, allowing it to work with any ordered key type and any value type.

**Note:** A sync-to-disk feature is currently under development, which will allow persisting the B+Tree data structure to disk for durability.

## Usage
Here is a basic example of how to use the library:

```go
package main

import (
    "fmt"
    "github.com/moshenahmias/bp3"
)

func main() {
    // Create a new B+Tree instance of order 3
    tree := bp3.New[int, string](3)

    // Insert elements
    tree.Insert(1, "one")
    tree.Insert(2, "two")
    tree.Insert(3, "three")

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
}
```

## Installation
To install the library, use `go get`:
```sh
go get github.com/moshenahmias/bp3
```
