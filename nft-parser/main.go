package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func main() {
	// Requires Admin - open raw volume, not a file
	f, err := os.Open(`\\.\C:`)
	if err != nil {
		log.Fatal(err) // Access denied? Run as Admin
	}
	defer f.Close()

	boot := make([]byte, 512)
	f.ReadAt(boot, 0)

	bytesPerSector := binary.LittleEndian.Uint16(boot[11:13])
	sectorsPerCluster := boot[13]
	mftCluster := binary.LittleEndian.Uint64(boot[48:56]) // 0x30 offset
	mftRecordSize := boot[64]                             // usually 0xF6 = -10 => 2^10 = 1024

	fmt.Printf("BytesPerSector: %d\n", bytesPerSector)
	fmt.Printf("SectorsPerCluster: %d\n", sectorsPerCluster)
	fmt.Printf("MFT starts at cluster: %d\n", mftCluster)
	fmt.Printf("MFT Record size flag: %d\n", int8(mftRecordSize))

	// Calculate byte offset of MFT
	clusterSize := int64(bytesPerSector) * int64(sectorsPerCluster)
	mftOffset := int64(mftCluster) * clusterSize

	// Read first MFT record (1024 bytes)
	mftRec := make([]byte, 1024)
	_, err = f.ReadAt(mftRec, mftOffset)
	if err != nil {
		log.Fatal(err)
	}

	// MFT signature must be "FILE"
	fmt.Printf("Signature: %s\n", string(mftRec[0:4])) // FILE
	// Check if resident
	if string(mftRec[0:4]) == "FILE" {
		fmt.Println("Got MFT #0 ($MFT itself) - parse attributes at offset",
			binary.LittleEndian.Uint16(mftRec[20:22]))
	}
}
