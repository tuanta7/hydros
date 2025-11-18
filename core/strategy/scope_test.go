package strategy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanta7/hydros/core"
)

func TestHierarchicScopeStrategy_ExactMatch(t *testing.T) {
	haystack := []string{"read", "write", "picture"}
	needles := []string{"read"}

	err := PrefixScopeStrategy(haystack, needles)
	assert.NoError(t, err)
}

func TestHierarchicScopeStrategy_PrefixMatch(t *testing.T) {
	haystack := []string{"picture", "profile"}
	needles := []string{"picture.read", "profile.view"}

	err := PrefixScopeStrategy(haystack, needles)
	assert.NoError(t, err)
}

func TestHierarchicScopeStrategy_MissingScope(t *testing.T) {
	haystack := []string{"picture"}
	needles := []string{"picture.read", "unknown"}

	err := PrefixScopeStrategy(haystack, needles)
	if assert.Error(t, err) {
		// ensure it's the library's invalid scope error type
		assert.Contains(t, err.Error(), core.ErrInvalidScope.Error())
	}
}

func TestHierarchicScopeStrategy_EmptyNeedles(t *testing.T) {
	haystack := []string{"picture"}
	needles := []string{}

	err := PrefixScopeStrategy(haystack, needles)
	assert.NoError(t, err)
}
