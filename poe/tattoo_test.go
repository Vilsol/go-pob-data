package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestTattooTables(t *testing.T) {
	initGameData(t, InitializePassiveSkillOverrides, InitializePassiveSkillOverrideTypes, InitializePassiveSkillTattoos, InitializePassiveSkillTattooTargetSets, InitializePassiveOverrideLimits, InitializeClientStrings, InitializeCurrencyExchanges)

	t.Run("PassiveSkillTattoos", func(t *testing.T) {
		testza.AssertGreater(t, len(PassiveSkillTattoos), 100)
		tattoo := PassiveSkillTattoos[0]
		testza.AssertEqual(t, "Ngamahu1", tattoo.GetOverride().ID)
		testza.AssertEqual(t, "Tattoo of the Ngamahu Firewalker", tattoo.GetTattoo().Name)
		testza.AssertNotNil(t, tattoo.GetSet())
		testza.AssertEqual(t, "1 Loyalty Tattoo", PassiveOverrideLimitsByID["Embrace"].Description)
		testza.AssertGreater(t, len(PassiveSkillOverrides), 300)
		testza.AssertGreater(t, len(PassiveSkillOverrideTypes), 10)
		testza.AssertEqual(t, 6, PassiveSkillOverridesByID["Ngamahu1"].StatValues[0])
	})

	t.Run("ClientStrings", func(t *testing.T) {
		testza.AssertGreater(t, len(ClientStrings), 5000)
		testza.AssertEqual(t, "Path of Exile", ClientStringsByID["GameName"].Text)
	})

	t.Run("CurrencyExchange", func(t *testing.T) {
		testza.AssertGreater(t, len(CurrencyExchanges), 500)
		testza.AssertNotNil(t, CurrencyExchanges[0].GetItem())
	})
}
