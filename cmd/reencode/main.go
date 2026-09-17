package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/raw"
)

func reencode[T any, P interface {
	*T
	msgp.Encodable
	msgp.Decodable
}](dir, name string) {
	f, err := os.Open(dir + "/" + name + ".json.gz")
	check(err)
	gz, err := gzip.NewReader(f)
	check(err)
	b, err := io.ReadAll(gz)
	check(err)
	var rows []P
	check(json.Unmarshal(b, &rows))

	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	w := msgp.NewWriter(bw)
	check(w.WriteInt64(int64(len(rows))))
	for _, r := range rows {
		check(r.EncodeMsg(w))
	}
	check(w.Flush())
	check(bw.Close())

	// Round-trip the way loader.LoadRaw reads it: count, then exactly that many objects, then EOF.
	rd := msgp.NewReader(brotli.NewReader(bytes.NewReader(buf.Bytes())))
	n, err := rd.ReadInt64()
	check(err)
	for i := int64(0); i < n; i++ {
		var t T
		check(P(&t).DecodeMsg(rd))
	}
	if _, err := rd.R.ReadByte(); !errors.Is(err, io.EOF) {
		panic("trailing bytes in " + name)
	}
	check(os.WriteFile(dir+"/"+name+".msgpack.br", buf.Bytes(), 0o755))
	fmt.Println(name, n, "rows", buf.Len(), "bytes")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Rebuilds a version's msgpack.br from its committed json.gz, for tables whose raw struct gained
// columns after extraction. Usage: go run ./cmd/reencode data/<version>/raw
func main() {
	dir := os.Args[1]
	reencode[raw.AlternateTreeVersion](dir, "AlternateTreeVersions")
	reencode[raw.PassiveSkill](dir, "PassiveSkills")
}
