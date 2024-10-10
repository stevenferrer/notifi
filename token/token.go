package token

type (
	ID    string
	CBKey string
)

type Token struct {
	// ID is the token id and acts as an API key
	ID ID `json:"api_key"`
	// CBKey is the callback key for validating on customer-end
	CBKey CBKey `json:"cb_key"`
}
