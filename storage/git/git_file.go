package git

import (
	"bytes"
	"io"
	"os"
)

// readFile reads a file in the local working tree.
func (gitSession *gitSession) readFile(path string) ([]byte, error) {
	var buf []byte

	file, err := gitSession.fs.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// writeFile write this buf to the file in the local working tree.
// Either new file will be created or existing one gets overwritten.
func (gitSession *gitSession) writeFile(path string, buf []byte) error {
	file, err := gitSession.fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(buf)); err != nil {
		return err
	}

	return nil
}
