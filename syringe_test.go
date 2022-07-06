package PhylumSyringGitlab

import (
	"reflect"
	"testing"
)

func TestNewSyringe(t *testing.T) {
	type args struct {
		gitlabToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *Syringe
		wantErr bool
	}{
		{"one", args{"bs8FExie7XVsVV7YbnG6"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSyringe(tt.args.gitlabToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSyringe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(&Syringe{}) {
				t.Errorf("NewSyringe() got = %v, want %v", got, tt.want)
			}
		})
	}
}
