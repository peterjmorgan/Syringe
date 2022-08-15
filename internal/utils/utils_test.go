package utils

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
	"testing"
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
			"vcsToken":    "bs8FExie7XVsVV7YbnG6",
			"phylumToken": "eyJhbGciOiJIUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIwZGVjZmE4OC0wMjNmLTQ3N2ItOGJiMy0yZmY5ODczMzdlYzMifQ.eyJleHAiOjE5NzE5MTY3NDgsImlhdCI6MTY1NzM3OTA1MCwianRpIjoiOWFkZWJiOWItYzdlYS00OGM0LTlkMzgtOTFmZDRhODVmMjBhIiwiaXNzIjoiaHR0cHM6Ly9sb2dpbi5waHlsdW0uaW8vYXV0aC9yZWFsbXMvcGh5bHVtIiwiYXVkIjoiaHR0cHM6Ly9sb2dpbi5waHlsdW0uaW8vYXV0aC9yZWFsbXMvcGh5bHVtIiwic3ViIjoiMDBiMjhhNTAtMWNiMS00YmJlLWJkMjQtZjBjNjNlNzljYzdkIiwidHlwIjoiT2ZmbGluZSIsImF6cCI6InBoeWx1bV9jbGkiLCJzZXNzaW9uX3N0YXRlIjoiOGUzYjU5MzMtZjI2Zi00NWQwLWIxMTQtYmJjNjk2MzI3YzdmIiwic2NvcGUiOiJvcGVuaWQgZW1haWwgb2ZmbGluZV9hY2Nlc3MgcHJvZmlsZSIsInNpZCI6IjhlM2I1OTMzLWYyNmYtNDVkMC1iMTE0LWJiYzY5NjMyN2M3ZiJ9.0f8vEO0gGKZgZV2XzQRmpbcWsoQiOQhFbll0A1Pd2ZM",
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
