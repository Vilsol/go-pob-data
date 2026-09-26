package raw

import (
	"bytes"
	"math/rand/v2"
	"reflect"
	"strconv"
	"testing"

	"github.com/tinylib/msgp/msgp"
)

func TestIndexHandlersEncodeDeterministic(t *testing.T) {
	keys := make([]string, 200)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	var want []byte
	for range 20 {
		rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
		handlers := IndexHandlers{}
		for _, k := range keys {
			handlers[k] = "handler_" + k
		}
		lt := LangTranslation{IndexHandlers: handlers, String: "x"}

		var buf bytes.Buffer
		w := msgp.NewWriter(&buf)
		if err := lt.EncodeMsg(w); err != nil {
			t.Fatal(err)
		}
		if err := w.Flush(); err != nil {
			t.Fatal(err)
		}
		marshalled, err := lt.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes(), marshalled) {
			t.Fatal("EncodeMsg and MarshalMsg disagree")
		}

		if want == nil {
			want = marshalled
		} else if !bytes.Equal(want, marshalled) {
			t.Fatal("the same map encoded to different bytes")
		}

		var decoded LangTranslation
		if err := decoded.DecodeMsg(msgp.NewReader(bytes.NewReader(marshalled))); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, lt) {
			t.Fatal("decoded translation differs")
		}
	}
}
