package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/in3rsha/bitcoin-utxo-dump/bitcoin/btcleveldb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
	defaultfolder := fmt.Sprintf("%s/.bitcoin/chainstate/", os.Getenv("HOME"))

	chainstate := flag.String("db", defaultfolder, "Location of bitcoin chainstate db.")
	flag.Parse()

	// check chainstate LevelDB folder exists
	if _, err := os.Stat(*chainstate); os.IsNotExist(err) {
		fmt.Println("Couldn't find", *chainstate)
		return
	}

	db, err := leveldb.OpenFile(*chainstate, &opt.Options{
		Compression: opt.NoCompression,
	})
	if err != nil {
		fmt.Println("Couldn't open LevelDB.")
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Iterate over LevelDB keys
	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		// first byte in key indicates the type of key we've got for leveldb
		prefix := key[0]

		// utxo entry
		if prefix == 67 { // 67 = 0x43 = C = "utxo"

			// ---
			// Key
			// ---

			//      430000155b9869d56c66d9e86e3c01de38e3892a42b99949fe109ac034fff6583900
			//      <><--------------------------------------------------------------><>
			//      /                               |                                  \
			//  type                          txid (little-endian)                      index (varint)

			// txid
			txidLE := key[1:33] // little-endian byte order

			// txid - reverse byte order
			txid := make([]byte, 0)                 // create empty byte slice (dont want to mess with txid directly)
			for i := len(txidLE) - 1; i >= 0; i-- { // run backwards through the txid slice
				txid = append(txid, txidLE[i]) // append each byte to the new byte slice
			}
			fmt.Println(hex.EncodeToString(txid)) // add to output results map

			// vout
			index := key[33:]

			// convert varint128 index to an integer
			vout := btcleveldb.Varint128Decode(index)
			fmt.Printf("%d\n", vout)
		}
	}
}
