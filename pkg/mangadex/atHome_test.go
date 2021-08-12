package mangadex

import (
	"fmt"
	"testing"
)

func TestHome(t *testing.T) {
	c := NewClient()

	fmt.Println(c.GetHomeUrl("e46e5118-80ce-4382-a506-f61a24865166"))
	fmt.Println(c.GetHomeUrl("1b34444b-0e93-4f68-bc72-e3b01099cf2f"))
	fmt.Println(c.GetHomeUrl("528d2e34-f5c6-4f58-8c4b-bd805500257e"))
}
