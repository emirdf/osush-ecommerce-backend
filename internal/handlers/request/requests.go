package request

type Dictionary struct {
	Tm string `json:"tm"`
	Ru string `json:"ru"`
}

type DictionaryMap struct {
	Tm map[string]interface{} `json:"tm"`
	Ru map[string]interface{} `json:"ru"`
}
