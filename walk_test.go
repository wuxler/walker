package walker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkDir(t *testing.T) {

	regulars := []string{}

	err := New().Check(IsRegular()).OnVisit(func(f File) error {
		regulars = append(regulars, trimRoot(f.Path()))
		return nil
	}).WalkDir(scaffolingRoot)

	assert.NoError(t, err)

	expects := []string{}
	for _, entry := range fileEntries {
		expects = append(expects, entry.Name())
	}
	assert.ElementsMatch(t, expects, regulars)
}
