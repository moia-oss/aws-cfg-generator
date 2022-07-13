package util

import (
	"reflect"
	"testing"
)

func TestOrderProfiles(t *testing.T) {

	tests := []struct {
		name              string
		unorderedProfiles []Profile
		want              []Profile
	}{
		{name: "mixed profile list",
			unorderedProfiles: []Profile{
				{ProfileName: "bar.int"},
				{ProfileName: "tools"},
				{ProfileName: "foo.dev"},
				{ProfileName: "cookies"},
				{ProfileName: "bar.prd"},
				{ProfileName: "gears"},
			},
			want: []Profile{
				{ProfileName: "cookies"},
				{ProfileName: "gears"},
				{ProfileName: "tools"},
				{ProfileName: "bar.int"},
				{ProfileName: "bar.prd"},
				{ProfileName: "foo.dev"},
			}},
		{name: "only single profiles",
			unorderedProfiles: []Profile{
				{ProfileName: "bar"},
				{ProfileName: "tools"},
				{ProfileName: "foo"},
				{ProfileName: "cookies"},
				{ProfileName: "bikes"},
				{ProfileName: "gears"},
			},
			want: []Profile{
				{ProfileName: "bar"},
				{ProfileName: "bikes"},
				{ProfileName: "cookies"},
				{ProfileName: "foo"},
				{ProfileName: "gears"},
				{ProfileName: "tools"},
			}},
		{name: "only staged profiles",
			unorderedProfiles: []Profile{
				{ProfileName: "bar.int"},
				{ProfileName: "tools.prd"},
				{ProfileName: "foo.dev"},
				{ProfileName: "cookies.dev"},
				{ProfileName: "bar.prd"},
				{ProfileName: "gears.poc"},
			},
			want: []Profile{
				{ProfileName: "bar.int"},
				{ProfileName: "bar.prd"},
				{ProfileName: "cookies.dev"},
				{ProfileName: "foo.dev"},
				{ProfileName: "gears.poc"},
				{ProfileName: "tools.prd"},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OrderProfiles(tt.unorderedProfiles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrderProfiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isStageProfile(t *testing.T) {

	tests := []struct {
		name    string
		profile Profile
		want    bool
	}{
		{name: "unique account",
			profile: Profile{ProfileName: "cookies"},
			want:    false},
		{name: "dev account",
			profile: Profile{ProfileName: "cookies.dev"},
			want:    true},
		{name: "global account",
			profile: Profile{ProfileName: "tools.global"},
			want:    false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStageProfile(tt.profile); got != tt.want {
				t.Errorf("isStageProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}
