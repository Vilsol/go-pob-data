// Package legacyraw is a frozen copy of raw and its map-encoding msgp code from before raw switched
// to tuples. It decodes the legacy raw/*.msgpack.br and stat_translations/*.msgpack.br assets, which
// are no longer regenerated, so tests can prove the tuple assets carry the same rows. Never edit it.
package legacyraw
