package extractor

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"path"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/pkg/errors"
)

var patchServer = map[Game]string{
	PoE1: "patch.pathofexile.com:12995",
	PoE2: "patch.pathofexile2.com:13060",
}

// patchReplyStrings is where the patch server's reply starts listing CDN URLs, after its opcode and 32 bytes.
const patchReplyStrings = 33

// LatestPatch asks game's patch server for the live patch, e.g. 3.29.3.3 or 0.5.5.4.2.
func LatestPatch(game Game) (string, error) {
	conn, err := net.DialTimeout("tcp", patchServer[game], 10*time.Second)
	if err != nil {
		return "", errors.Wrap(err, "could not reach the patch server")
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return "", errors.Wrap(err, "could not set a deadline")
	}
	if _, err := conn.Write([]byte{1, 6}); err != nil {
		return "", errors.Wrap(err, "could not query the patch server")
	}

	reply := make([]byte, 4096)
	n, err := io.ReadAtLeast(conn, reply, patchReplyStrings+2)
	if err != nil {
		return "", errors.Wrap(err, "could not read the patch server's reply")
	}
	return parsePatchReply(reply[:n])
}

// parsePatchReply reads the first CDN URL, a big-endian u16 length in UTF-16 units and then UTF-16LE text,
// and returns its last path element.
func parsePatchReply(reply []byte) (string, error) {
	if len(reply) < patchReplyStrings+2 {
		return "", fmt.Errorf("patch server reply of %d bytes is too short", len(reply))
	}
	n := int(binary.BigEndian.Uint16(reply[patchReplyStrings:]))
	text := reply[patchReplyStrings+2:]
	if n == 0 || len(text) < 2*n {
		return "", fmt.Errorf("patch server reply holds no URL of %d units", n)
	}

	units := make([]uint16, n)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(text[2*i:])
	}
	u, err := url.Parse(string(utf16.Decode(units)))
	if err != nil {
		return "", errors.Wrap(err, "patch server sent a bad URL")
	}

	patch := path.Base(strings.TrimSuffix(u.Path, "/"))
	if u.Host == "" || patch == "." || patch == "/" {
		return "", fmt.Errorf("patch server sent %q, which names no patch", u.String())
	}
	return patch, nil
}
