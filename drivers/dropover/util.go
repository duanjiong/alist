package dropover

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
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
