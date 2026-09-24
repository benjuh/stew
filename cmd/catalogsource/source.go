package catalogsource

import (
	"fmt"

	"github.com/benjuh/stew/catalog"
	"github.com/spf13/viper"
)

func Load(url string) (catalog.Index, error) {
	return LoadWithOptions(url, false)
}

func LoadWithOptions(url string, refresh bool) (catalog.Index, error) {
	if url == "" {
		url = viper.GetString("catalogURL")
	}
	if url == "" {
		return catalog.Index{}, fmt.Errorf("no catalog configured; use --catalog-url or set catalogURL")
	}
	return catalog.LoadCachedWithOptions(url, viper.GetString("templatesPath"), refresh)
}
