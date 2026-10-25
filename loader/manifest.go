package loader

import (
	"context"
	"encoding/json"
	"slices"
	"sync"

	"github.com/pkg/errors"
)

// Manifest lists what a version publishes, so a consumer can tell a table the version does not have
// from a fetch that failed. The extractor writes it to <version>/manifest.json.
type Manifest struct {
	// Tables are the raw tables by name, such as "Mods".
	Tables []string `json:"tables"`
	// StatTranslations are the stat description files by name, such as "stat_descriptions".
	StatTranslations []string `json:"statTranslations"`
	// Languages are the stat translation language directories, such as "en".
	Languages []string `json:"languages"`
	// Files are the game text files LoadFile reads, by game path.
	Files []string `json:"files"`
	// JSONOnly is set when the tables have no tuple assets, only json.gz and json.br, so LoadRaw cannot
	// read them. Stat translations are unaffected.
	JSONOnly bool `json:"jsonOnly,omitempty"`
}

func (m *Manifest) HasTable(name string) bool {
	return slices.Contains(m.Tables, name)
}

func (m *Manifest) HasStatTranslation(name string) bool {
	return slices.Contains(m.StatTranslations, name)
}

func (m *Manifest) HasFile(path string) bool {
	return slices.Contains(m.Files, path)
}

// ManifestURL is the URL LoadManifest reads version's manifest from.
func ManifestURL(version string) string {
	return cdnBase + version + "/manifest.json"
}

var (
	manifests     = map[string]*Manifest{}
	manifestsLock sync.Mutex
)

// noCache keeps the manifest out of the asset cache: a version gains tables when it is re-extracted,
// so a cached manifest would hide them for good.
type noCache struct{}

func (noCache) Get(string) ([]byte, error) { return nil, nil }
func (noCache) Set(string, []byte) error   { return nil }
func (noCache) Exists(string) bool         { return false }

// LoadManifest fetches version's manifest once per process. A failed fetch is returned, not remembered.
func LoadManifest(ctx context.Context, version string) (*Manifest, error) {
	manifestsLock.Lock()
	defer manifestsLock.Unlock()

	if m, ok := manifests[version]; ok {
		return m, nil
	}

	b, err := fetchAsset(ctx, ManifestURL(version), noCache{})
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Wrap(err, "failed to decode manifest of "+version)
	}

	manifests[version] = &m
	return &m, nil
}

// HasTable reports whether version publishes the raw table name. An error means it could not be told.
func HasTable(ctx context.Context, version string, name string) (bool, error) {
	m, err := LoadManifest(ctx, version)
	if err != nil {
		return false, err
	}
	return m.HasTable(name), nil
}
