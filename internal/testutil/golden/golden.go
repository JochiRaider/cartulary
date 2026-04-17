package golden

import (
	"os"
	"path/filepath"
	"runtime"
)

func Dir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func Path(parts ...string) string {
	segments := append([]string{Dir()}, parts...)
	return filepath.Join(segments...)
}

func Read(parts ...string) ([]byte, error) {
	return os.ReadFile(Path(parts...))
}

func MustRead(parts ...string) []byte {
	data, err := Read(parts...)
	if err != nil {
		panic(err)
	}
	return data
}
