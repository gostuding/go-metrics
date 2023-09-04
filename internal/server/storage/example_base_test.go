package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Default vars for use in tests. Connection string for SQLStore's test is default.
// Corrects it for your database setings.
var (
	defFileName    = filepath.Join(os.TempDir(), "Memory.strg")
	restoreStorage = false
	saveInterval   = 300
	gType          = "gauge"
	cType          = "counter"
	defMetricName  = "mName"
	memStorage     *MemStorage     // Storage used in MemStorage's tests
	ctx            context.Context // default context
)

func Example() {
	// Create new Memory Storage example. For restore storage use 'restoreStorage' as true.
	mem, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	if err != nil {
		fmt.Printf("Create memory storage error: %v", err)
		return
	}
	// Do any actions with storage...
	// ...
	// ...
	// Save storage before work finish. Stop func do the same.
	err = mem.Save()
	if err != nil {
		fmt.Printf("Save memory storage error: %v", err)
	} else {
		fmt.Println("Memory storage save success")
	}

	// Output:
	// Memory storage save success
}
