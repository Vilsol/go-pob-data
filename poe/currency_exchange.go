package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type CurrencyExchange struct {
	raw.CurrencyExchange
}

var CurrencyExchanges []*CurrencyExchange

func InitializeCurrencyExchanges(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "CurrencyExchange", &CurrencyExchanges, nil, assetCache)
}

func (c *CurrencyExchange) GetItem() *BaseItemType {
	return BaseItemTypes[c.Item]
}
