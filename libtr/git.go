package libtr

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/src-d/go-git.v4"
)

var gitVersion string

func getGitVersion() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Panicln(err)
	}
	repo, err := git.PlainOpen(dir)
	if err != nil {
		gitVersion = "build"
		return gitVersion
	}
	head, err := repo.Head()
	if err != nil {
		log.Panicln(err)
	}
	gitVersion = fmt.Sprint(head.Hash())
	return gitVersion
}
