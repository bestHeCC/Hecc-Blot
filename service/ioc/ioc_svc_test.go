package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type iInterface interface {
	Test() string
}

type derive struct{}

func (m derive) Test() string {
	return "set test"
}

type defaultTest struct {
	One iInterface `inject:""`
}

type customTest struct {
	One iInterface `inject:"custom"`
}

type composeTest struct {
	defaultTest

	Child iInterface `inject:"custom"`
}

func TestIocSvc(t *testing.T) {
	Set(new(iInterface), derive{})
	SetWithName(new(iInterface), "custom", derive{})

	t.Run("default", func(t *testing.T) {
		var d1 defaultTest
		Inject(&d1)

		assert.Equal(t, "set test", d1.One.Test())
	})

	t.Run("custom", func(t *testing.T) {
		var d2 customTest
		Inject(&d2)

		assert.Equal(t, "set test", d2.One.Test())

	})

	t.Run("compose", func(t *testing.T) {
		d3 := composeTest{}
		Inject(&d3)

		assert.Equal(t, "set test", d3.One.Test())
		assert.Equal(t, "set test", d3.defaultTest.One.Test())
		assert.Equal(t, "set test", d3.Child.Test())
	})
}
