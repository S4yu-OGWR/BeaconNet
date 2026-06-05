# BeaconNet

A Golang library for manipulating binary streams specific to the Minecraft 1.8.x protocol.

### Features
* **VarInt & VarLong:** Full implementation according to the Minecraft protocol specification.
* **String Handling:** Reads and writes UTF-8 strings with a length prefix.
* **Low Level:** Runs on top of `io.Reader` and `io.Writer` (compatible with TCP Conn, File, or Buffer).

### Installation
```bash
go get github.com/S4yu-OGWR/BeaconNet
```

### Usage Example (Quick Start)
This is the most important part. Provide a code example that can be directly copied and run.
### 🛠 Usage Example

```go
package main

import (
	"fmt"

	"github.com/S4yu-OGWR/BeaconNet/packet"
)

func main() {
	compressor := packet.NewCompressor(256)

	fmt.Println("=== Test 1: Small Packet ===")
	smallPacketData := packet.NewWriter().String("Hi").Raw()
	compressed1, _ := compressor.CompressPacket(0x02, smallPacketData)
	fmt.Printf("Small packet size: %d bytes (tidak dikompres karena < 256)\n", len(compressed1))

	fmt.Println("\n=== Test 2: Large Packet ===")
	largeData := packet.NewWriter().String("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris. memekurun dadwa kontol gede sekali sangat besar sekali sampai aku kelelahan bekerja demi seorang waifu").Raw()
	originalSize := len(packet.NewWriter().VarInt(0x02).Bytes(largeData).Raw())
	compressed2, _ := compressor.CompressPacket(0x02, largeData)

	fmt.Printf("Original size: %d bytes\n", originalSize)
	fmt.Printf("Compressed size: %d bytes\n", len(compressed2))
	fmt.Printf("Compression ratio: %.2f%%\n", float64(len(compressed2))/float64(originalSize)*100)

	fmt.Println("\n=== Test 3: Decompress Verify ===")
	decodeID, decodeData, _ := compressor.DecompressPacket(compressed2)
	r := packet.NewReader(decodeData)
	decodeMessage := r.String()

	fmt.Printf("Decode Packet ID: 0x%02x\n", decodeID)
	fmt.Printf("Decode message: %s\n", decodeMessage)
}

```
[![Go Reference](https://pkg.go.dev/badge/github.com/Kartendsy/mc1.8.x-bin.svg)](https://pkg.go.dev/github.com/Kartendsy/mc1.8.x-bin)
