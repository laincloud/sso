package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDatabaseShouldWork(t *testing.T) {
	th := NewTestHelper(t)
	assert.NotPanics(th.T, func() { testBack.(*UserBack).InitDatabase() })
}
