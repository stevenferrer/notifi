package token

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

func NewID() ID {
	return ID(genUUID())
}

func NewCBKey() CBKey {
	cbKey := base64.RawURLEncoding.EncodeToString([]byte(genUUID()))
	return CBKey(cbKey)
}

func genUUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
