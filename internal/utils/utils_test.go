package utils

import (
	"os"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func setupEnv(t *testing.T, envVarName string) func(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Failed to load .env for testing: %v\n", err)
	}

	return func(t *testing.T) {
		err := os.Unsetenv(envVarName)
		if err != nil {
			log.Fatalf("Failed to unset %v: %v\n", envVarName, err)
		}
	}
}

func TestReadEnvironment(t *testing.T) {
	tearDown := setupEnv(t, "SYRINGE_VCS_TOKEN_GITLAB")
	defer tearDown(t)

	tests := []struct {
		name    string
		want    map[string]string
		wantErr bool
	}{
		{"one", map[string]string{
			"vcs":         "gitlab",
			"vcsToken":    "",
			"phylumToken": "",
			"phylumGroup": "petetest1",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadEnvironment()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadEnvironment() got = %v, want %v", got, tt.want)
			}
		})
	}
}
