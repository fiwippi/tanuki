package pretty

import (
	"fmt"
	"reflect"
	"strings"
)

// MapKeys returns a string representation of all map keys
func MapKeys(a interface{}) string {
	keys := reflect.ValueOf(a).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	return fmt.Sprintf("[%v]", strings.Join(strkeys, ", "))
}
