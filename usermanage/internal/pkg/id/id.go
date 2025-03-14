package id

import (
	"strings"

	"github.com/google/uuid"
)

// GenerateUUID generates a UUID.
//
// noDash: whether to remove the dash (default is false).
//
// Examples
// 
//	id.GenerateUUID() // "f47ac10b-58cc-4372-a567-0e02b2c3d479"
//	id.GenerateUUID(true) // "f47ac10b58cc4372a5670e02b2c3d479"
func GenerateUUID(noDash ...bool) string {
	uuid := uuid.New().String()
	if len(noDash) > 0 && noDash[0] {
		return strings.Replace(uuid, "-", "", -1)
	} else {
		return uuid
	}
}
