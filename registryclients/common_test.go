package registryclients

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func Test_getLatestTag(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "semver tags",
			tags: []string{"1.0.0", "2.0.0", "1.1.0"},
			want: "2.0.0",
		},
		{
			name: "non-semver tags",
			tags: []string{"latest", "stable", "dev"},
			want: "dev",
		},
		{
			name: "mixed tags",
			tags: []string{"1.0.0", "latest", "2.0.0"},
			want: "2.0.0",
		},
		{
			name: "no tags",
			tags: []string{},
			want: "",
		},
		{
			name: "only invalid semver tags",
			tags: []string{"invalid", "not.valid"},
			want: "not.valid",
		},
		{
			name: "Semver with hyphens",
			tags: []string{"1.0.0-alpha", "1.0.0-beta", "1.0.0"},
			want: "1.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestTag(tt.tags); got != tt.want {
				t.Errorf("getLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("check semver sorting", func(t *testing.T) {
		tags := []string{"1.1.0", "1.0.0", "2.0.0"}
		expectedVersions := []*semver.Version{}
		for _, tag := range tags {
			v, _ := semver.NewVersion(tag)
			expectedVersions = append(expectedVersions, v)
		}

		versions := []*semver.Version{}
		nonSemverTags := []string{}

		for _, tag := range tags {
			version, err := semver.NewVersion(tag)
			if err == nil {
				versions = append(versions, version)
			} else {
				nonSemverTags = append(nonSemverTags, tag)
			}
		}

		if len(versions) != len(expectedVersions) {
			t.Errorf("length of versions is incorrect")
		}

		for i := range versions {
			if versions[i].String() != expectedVersions[i].String() {
				t.Errorf("version at index %d is incorrect", i)
			}
		}

		if len(nonSemverTags) != 0 {
			t.Errorf("nonSemverTags should be empty")
		}
	})

	t.Run("check non-semver sorting", func(t *testing.T) {
		tags := []string{"latest", "stable", "dev"}
		expectedNonSemverTags := []string{"latest", "stable", "dev"}

		versions := []*semver.Version{}
		nonSemverTags := []string{}

		for _, tag := range tags {
			version, err := semver.NewVersion(tag)
			if err == nil {
				versions = append(versions, version)
			} else {
				nonSemverTags = append(nonSemverTags, tag)
			}
		}

		if len(versions) != 0 {
			t.Errorf("versions should be empty")
		}

		if !reflect.DeepEqual(nonSemverTags, expectedNonSemverTags) {
			t.Errorf("nonSemverTags should be equal")
		}
	})
}
