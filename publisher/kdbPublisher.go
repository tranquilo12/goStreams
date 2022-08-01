package publisher

import (
	"fmt"
	kdb "github.com/sv/kdbgo"
	"lightning/utils/db"
)

// CreateKdbConn Create a new KDB connection
func CreateKdbConn() *kdb.KDBConn {
	// Create a new kdb connection
	kdbConn, err := kdb.DialKDB("localhost", 5000, "")
	db.CheckErr(err)

	res, err := kdbConn.Call("til", kdb.Int(10))
	if err != nil {
		fmt.Println("Query failed:", err)
	}
	fmt.Println("Connection Successful, Result:", res)

	return kdbConn
}

//
