package loader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tinylib/msgp/msgp"
)

const cdnBase = "https://go-pob-data.pages.dev/data/%s/raw/%s.msgpack.br"

// LoadRaw loads a raw brotli-compressed json dump from remote source
//
// Returns data from cache if found
func LoadRaw[T msgp.Decodable](version string, name string, onInit func(count int64), assetCache AssetCache, hooks ...func(obj T)) ([]T, error) {
	url := fmt.Sprintf(cdnBase, version, name)

	log.Debug().Str("version", version).Str("name", name).Msg("loading raw asset")

	var b []byte
	if assetCache.Exists(url) {
		get, err := assetCache.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve url from cache: "+url)
		}
		b = get
	} else {
		log.Debug().Str("url", url).Msg("fetching")
		response, err := http.DefaultClient.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch url: "+url)
		}
		defer response.Body.Close()

		b, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		defer func() {
			_ = assetCache.Set(url, b)
		}()
	}

	unzipStream := brotli.NewReader(bytes.NewReader(b))

	r := msgp.NewReader(unzipStream)
	elementCount, _ := r.ReadInt64()
	if onInit != nil {
		onInit(elementCount)
	}

	out := make([]T, elementCount)
	if elementCount > 0 {
		elemType := reflect.ValueOf(out[0]).Type().Elem()
		for i := int64(0); i < elementCount; i++ {
			out[i] = reflect.New(elemType).Interface().(T)
			if err := out[i].DecodeMsg(r); err != nil {
				return nil, errors.Wrap(err, "failed to read message")
			}
			for _, hook := range hooks {
				hook(out[i])
			}
		}
	}

	return out, nil
}

func InitHelper[T msgp.Decodable](version string, name string, target *[]T, onInit func(count int64), assetCache AssetCache, hooks ...func(obj T)) error {
	loadedRaw, err := LoadRaw[T](version, name, onInit, assetCache, hooks...)
	if err != nil {
		return err
	}

	*target = loadedRaw

	return nil
}
