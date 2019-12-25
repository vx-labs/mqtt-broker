package store

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkPut(b *testing.B) {
	dbFile := fmt.Sprintf("%s/db.bolt", os.TempDir())
	defer os.RemoveAll(dbFile)
	db, err := New(Options{Path: dbFile})
	require.NoError(b, err)
	for i := 0; i < b.N; i++ {
		db.Put([]byte("a"), []byte("a"), 0, uint64(i))
	}
}
