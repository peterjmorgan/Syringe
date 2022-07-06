package PhylumSyringGitlab

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"io"
	"os"
)

func init() {
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: false,
		DisableTimestamp:       false,
	})
	logFile, err := os.OpenFile("LOG_PhylumSyringeGitlab.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: failed to open logfile")
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetLevel(log.DebugLevel)
}

func NewSyringe(gitlabToken string) (*Syringe, error) {
	gitlabClient, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
		return nil, err
	}

	return &Syringe{Gitlab: gitlabClient}, nil
}
