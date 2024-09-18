package biz

import (
	"fmt"
	"testing"
)

func TestBase62(t *testing.T) {
	var id int64 = 1
	fmt.Println(characters)
	t.Log(base62Encode(id))
	fmt.Println(characters)

}
