package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func calculateSHA256(reader io.Reader) (string, io.Reader) {
	data := bytes.NewBuffer(nil)
	hash := sha256.New()
	for {
		buf := make([]byte, 4096)
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil
		}
		data.Write(buf[:n])
		hash.Write(buf[:n])
	}

	return hex.EncodeToString(hash.Sum(nil)), data
}

func main() {
	// 使用字符串作为示例，你也可以使用其他实现了 io.Reader 接口的类型
	dataFile, _ := os.Open("/Users/duanjiong/Downloads/Medis.dmg")

	sha256Value, _ := calculateSHA256(dataFile)

	fmt.Println("SHA-256 value:", sha256Value)
}
