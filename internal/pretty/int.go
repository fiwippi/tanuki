package pretty

import (
	"fmt"
	"strings"
)

func Padded(num int, p int) string {
	var s = fmt.Sprintf("%d", num)
	var zeroNum = p - len(s)
	var sb strings.Builder

	for i := 0; i < zeroNum; i++ {
		sb.WriteRune('0')
	}
	sb.WriteString(s)

	return sb.String()
}
