package extractor

import (
	"encoding/json"
	"io"
	"io/fs"
	"strings"

	"github.com/oriath-net/pogo/dat"
	"github.com/pkg/errors"
)

var parser *dat.DataParser

func LoadParser(gameVersion string) {
	LoadSchema(gameVersion)
	parser = dat.InitParser(gameVersion, &schemaFS{})
}

type schemaFS struct{}

func (s *schemaFS) Open(name string) (fs.File, error) {
	cleanName := strings.Split(name, ".")[0]

	format := tableMap[cleanName].ToJSONFormat()
	if format.File == "" {
		format.File = cleanName
	}

	data, err := json.Marshal(format)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal data")
	}

	return &schemaFSFile{Data: data}, nil
}

type schemaFSFile struct {
	Data   []byte
	Offset int
}

func (s *schemaFSFile) Stat() (fs.FileInfo, error) {
	// Do nothing
	return nil, nil
}

func (s *schemaFSFile) Read(bytes []byte) (int, error) {
	copied := copy(bytes, s.Data[s.Offset:])
	s.Offset += copied

	if s.Offset >= len(s.Data) {
		return copied, io.EOF
	}

	return copied, nil
}

func (s *schemaFSFile) Close() error {
	// Do nothing
	return nil
}

func ParseDat(data io.Reader, filename string) ([]interface{}, error) {
	return parser.Parse(data, filename)
}
