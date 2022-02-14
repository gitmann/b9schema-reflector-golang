package reflector

import (
	"reflect"
	"testing"
)

func TestNewStructFieldTag(t *testing.T) {
	testCases := []struct {
		name    string
		tag     string
		wantTag *StructFieldTag
	}{
		{
			name: "nil",
		},
		{
			name:    "ignore",
			tag:     "-",
			wantTag: &StructFieldTag{Ignore: true},
		},
		{
			name:    "ignore, double-quote",
			tag:     `"-"`,
			wantTag: &StructFieldTag{Ignore: true},
		},
		{
			name:    "ignore, single-quote",
			tag:     `'-'`,
			wantTag: &StructFieldTag{Ignore: true},
		},
		{
			name:    "ignore, comma",
			tag:     `"-,"`,
			wantTag: &StructFieldTag{Alias: "-"},
		},
		{
			name:    "alias",
			tag:     `"abc"`,
			wantTag: &StructFieldTag{Alias: "abc"},
		},
		{
			name:    "alias, comma",
			tag:     `"abc,"`,
			wantTag: &StructFieldTag{Alias: "abc"},
		},
		{
			name:    "alias, options",
			tag:     `"abc,def,ghi"`,
			wantTag: &StructFieldTag{Alias: "abc", Options: "def,ghi"},
		},
		{
			name:    "comma, options",
			tag:     `",def,ghi"`,
			wantTag: &StructFieldTag{Options: "def,ghi"},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			gotTag := NewStructFieldTag(test.tag)

			if !reflect.DeepEqual(gotTag, test.wantTag) {
				t.Errorf("TEST_FAIL %s: got=%v want=%v", test.name, gotTag, test.wantTag)
			} else {
				t.Logf("TEST_OK %s: got=%v", test.name, gotTag)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	testCases := []struct {
		name     string
		tag      reflect.StructTag
		wantTags map[string]*StructFieldTag
	}{
		{
			name:     "nil",
			wantTags: Tags{},
		},
		{
			name: "1 tag",
			tag:  `json:"abc,def"`,
			wantTags: Tags{
				"json": &StructFieldTag{
					Alias:   "abc",
					Options: "def",
				},
			},
		},
		{
			name: "2 tags",
			tag:  `json:"abc,def" bigquery:"-,xyz,123"`,
			wantTags: Tags{
				"json": &StructFieldTag{
					Alias:   "abc",
					Options: "def",
				},
				"bigquery": &StructFieldTag{
					Alias:   "-",
					Options: "xyz,123",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			gotTags := ParseTags(test.tag)

			allKeys := map[string]interface{}{}
			for tagName, _ := range test.wantTags {
				allKeys[tagName] = nil
			}
			for tagName, _ := range gotTags {
				allKeys[tagName] = nil
			}

			for tagName, _ := range allKeys {
				got := gotTags[tagName]
				want := test.wantTags[tagName]
				if !reflect.DeepEqual(got, want) {
					t.Errorf("TEST_FAIL %s: %q got=%v want=%v", test.name, tagName, got, want)
				} else {
					t.Logf("TEST_OK %s: %q got=%v", test.name, tagName, got)
				}
			}
		})
	}
}
