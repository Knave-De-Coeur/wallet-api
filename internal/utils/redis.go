package utils

import (
	"fmt"
)

func GenerateRedisKey(id int) string {
	return fmt.Sprintf("%d-balance", id)
}
