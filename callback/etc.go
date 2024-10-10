package callback

import (
	"strings"

	"github.com/google/uuid"
)

func NewID() ID {
	return ID(genUUID())
}

func genUUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
