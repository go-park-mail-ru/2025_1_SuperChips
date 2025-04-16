package rest

import "strconv"

func parsePinID(idStr string) (uint64, error) {
	return strconv.ParseUint(idStr, 10, 64)
}
