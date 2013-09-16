package backup

import (
	"io"
)

type Storage interface {
	Set(key string, data []byte) error
	Get(key string, writer io.Writer) error
}
