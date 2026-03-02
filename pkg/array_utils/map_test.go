package arrayutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	users := []User{
		{Name: "Alice", Age: 21},
		{Name: "Bob", Age: 35},
		{Name: "Carol", Age: 27},
	}

	names := Map(users, func(u User) string {
		return u.Name
	})

	expected := []string{"Alice", "Bob", "Carol"}
	assert.Equal(t, names, expected)
}
