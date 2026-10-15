package extractor

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/oriath-net/pogo/dat"
	"github.com/pkg/errors"
)

var parser *dat.DataParser

func LoadParser(game Game, gameVersion string) {
	LoadSchema(game, gameVersion)
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
	b, err := io.ReadAll(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading dat file")
	}

	if err := checkRowWidth(b, filename); err != nil {
		return nil, err
	}

	result, err := parser.Parse(bytes.NewReader(b), filename)
	return result, errors.Wrap(err, "failed parsing dat file")
}

// datBoundary separates the fixed rows from the variable data; pogo finds it the same way.
var datBoundary = bytes.Repeat([]byte{0xbb}, 8)

// checkRowWidth refuses a table whose schema row width differs from the data's: pogo takes the data's
// width and decodes a wrong layout into wrong columns without an error.
func checkRowWidth(b []byte, filename string) error {
	name := strings.Split(path.Base(filename), ".")[0]
	if len(b) < 4 {
		return fmt.Errorf("table %s: %d bytes is too short for a row count", name, len(b))
	}

	rowCount := int(int32(binary.LittleEndian.Uint32(b)))
	if rowCount <= 0 {
		return nil
	}

	boundary := bytes.Index(b, datBoundary)
	if boundary < 4 {
		return fmt.Errorf("table %s: no data boundary", name)
	}

	schemaWidth := SchemaRowWidth(name, filename)
	fixed := boundary - 4
	if fixed%rowCount != 0 || fixed/rowCount != schemaWidth {
		return fmt.Errorf(
			"table %s: schema row width %d != data row width %d (%d fixed bytes / %d rows), schema from %s",
			name, schemaWidth, fixed/rowCount, fixed, rowCount, tableSource[name],
		)
	}

	return nil
}

// SchemaRowWidth is the row width pogo would decode table name with for filename's dat variant.
func SchemaRowWidth(name string, filename string) int {
	width := dat.Dat32
	if ext := path.Ext(filename); ext == ".dat64" || ext == ".datc64" {
		width = dat.Datc64
	}

	n := 0
	for _, field := range tableMap[name].ToJSONFormat().Fields {
		if ft := dat.FieldType(field.Type); ft.Valid() {
			n += ft.Size(width)
		}
	}
	return n
}
