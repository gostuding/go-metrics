// Package storage implements Storage interface.
// Two types of data storage are available:
//
// 1. Data storage from RAM;
//
// 2. Data storage in a database (postgresql).
//
// When choosing the first type of data storage, data is periodically saved to a file.
package storage
