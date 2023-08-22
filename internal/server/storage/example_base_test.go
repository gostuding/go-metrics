package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Default vars for use in tests.
var (
	defFileName    = filepath.Join(os.TempDir(), "Memory.strg")
	restoreStorage = false
	saveInterval   = 300
	dsnString      = "host=localhost user=postgres database=metrics" // Connection strin for SQLStore's test. Corrects it for your database.
	gType          = "gauge"
	cType          = "counter"
	defMetricName  = "mName"
	memStorage     *MemStorage     // Storage used in MemStorage's tests
	sqlStorage     *SQLStorage     // Storage used in SQLStorage's tests
	ctx            context.Context // default context
)

func Example() {
	//Create new Memory Storage example. For restore storage use 'restoreStorage' as true.
	mem, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	if err != nil {
		fmt.Printf("Create memory storage error: %v", err)
		return
	}
	// Do any actions with storage...
	//...
	//...
	// Save storage before work finish. Stop func do the same.
	err = mem.Save()
	if err != nil {
		fmt.Printf("Save memory storage error: %v", err)
	} else {
		fmt.Println("Memory storage save success")
	}

	// Create SQL Storage example.
	// Database structure is checking when NewSQLStorage is calling.
	sqlStrg, err := NewSQLStorage(dsnString)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	// Do any actions with storage...
	//...
	//...
	// Stop storage before work finish.
	err = sqlStrg.Stop()
	if err != nil {
		fmt.Printf("Stop sql storage error: %v", err)
	} else {
		fmt.Println("SQL storage stop success")
	}

	// Output:
	// Memory storage save success
	// SQL storage stop success
}
