package run

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskfileNestingWithDir(t *testing.T) {
	ts, err := Load("./testdata/very-nested")
	assert.NoError(t, err)

	testNesting(t, ts)

	assert.Equal(t, "testdata/very-nested",
		ts["test"].(*scriptTask).dir)
	assert.Equal(t, "testdata/very-nested/child",
		ts["child/test"].(*scriptTask).dir)
	assert.Equal(t, "testdata/very-nested/child/grandchild",
		ts["child/grandchild/test"].(*scriptTask).dir)
}

func TestTaskfileNestingWithDot(t *testing.T) {
	os.Chdir("testdata/very-nested")
	defer os.Chdir("../..")

	ts, err := Load(".")
	assert.NoError(t, err)

	testNesting(t, ts)

	assert.Equal(t, ".",
		ts["test"].(*scriptTask).dir)
	assert.Equal(t, "child",
		ts["child/test"].(*scriptTask).dir)
	assert.Equal(t, "child/grandchild",
		ts["child/grandchild/test"].(*scriptTask).dir)
}

func testNesting(t *testing.T, ts Tasks) {
	metas := map[string]TaskMetadata{}
	for id, t := range ts {
		metas[id] = t.Metadata()
	}

	assert.Equal(t, map[string]TaskMetadata{
		"test": {
			ID:           "test",
			Type:         "short",
			Dependencies: []string{"child/test"},
			Watch:        []string{"file"},
		},
		"child/test": {
			ID:           "child/test",
			Type:         "short",
			Dependencies: []string{"child/grandchild/test"},
			Watch:        []string{"child/file"},
		},
		"child/grandchild/test": {
			ID:    "child/grandchild/test",
			Type:  "short",
			Watch: []string{"child/grandchild/file"},
		},
	}, metas)
}

func TestDescriptions(t *testing.T) {
	ts, err := Load("./testdata/task-descriptions")
	assert.NoError(t, err)

	metas := map[string]TaskMetadata{}
	for id, t := range ts {
		metas[id] = t.Metadata()
	}

	fPath := "./testdata/task-descriptions/out.log"
	if _, err := os.Stat(fPath); os.IsNotExist(err) {
		// Expected output does not exist! Create it.
		err := os.WriteFile(fPath, []byte(fmt.Sprintf("%+v", metas)), 0644)
		require.NoError(t, err)
	}

	expected, err := os.ReadFile(fPath)
	require.NoError(t, err)
	
	dmp := diffmatchpatch.New()
	diff := dmp.DiffMain(string(expected), fmt.Sprintf("%+v", metas), false)
	if len(diff) != 1 {
		log.Printf("Unexpected output from task descriptions test, saved to fail.log:\n%s", dmp.DiffPrettyText(diff))
		errFilePath := "./testdata/task-descriptions/fail.log"
		err := os.WriteFile(errFilePath, []byte(fmt.Sprintf("%+v", metas)), 0644)
		require.NoError(t, err)
	}

}
