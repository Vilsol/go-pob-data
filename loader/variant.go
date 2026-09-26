package loader

import "context"

// VariantFunc picks the variant LoadRaw and LoadTranslation read an asset of version in, for an asset
// such as "raw/Mods" or "stat_translations/en/stat_descriptions". fields.MinVariant is one. When it
// returns false, the full tuple asset is read.
type VariantFunc func(version string, asset string) (string, bool)

type variantsKey struct{}

// WithVariants makes loads under ctx read the variants variants picks, e.g. the .min assets:
//
//	ctx = loader.WithVariants(ctx, fields.MinVariant)
func WithVariants(ctx context.Context, variants VariantFunc) context.Context {
	return context.WithValue(ctx, variantsKey{}, variants)
}

// VariantURL is the URL LoadRaw and LoadTranslation read asset of version from under ctx, for rows of
// row's type; a prefetch that requests it hits the same cache entry.
func VariantURL(ctx context.Context, version string, asset string, row any) string {
	if variants, ok := ctx.Value(variantsKey{}).(VariantFunc); ok {
		if variant, ok := variants(version, asset); ok {
			return AssetURL(version, asset, variant)
		}
	}
	return AssetURL(version, asset, TupleVariant(row))
}
