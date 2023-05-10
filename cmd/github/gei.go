package github

import (
	"errors"
	"log"
	"math"
	"os/exec"
	"time"
)

func MigrateCodeScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	cmd := exec.Command("gh", "gei", "migrate-code-scanning-alerts", "--source-repo", repository,
		"--source-org", sourceOrg, "--target-org", targetOrg, "--github-source-pat", sourceToken,
		"--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Println("failed to migrate code scanning alerts: ", err)
		return err
	}

	return nil
}

func MigrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	cmd := exec.Command("gh", "gei", "migrate-secret-alerts", "--source-repo", repository, "--source-org", sourceOrg, "--target-org", targetOrg, "--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Println("failed to migrate secret scanning remediations: ", err)
		return err
	}

	return nil
}

func MigrateRepo(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	cmd := exec.Command("gh", "gei", "migrate-repo", "--source-repo",
		repository, "--github-source-org", sourceOrg, "--github-target-org", targetOrg,
		"--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Println("failed to migrate repository: ", err)
		return err
	}

	//call exponentialBackoff
	err = exponentialBackoff(func() (Repository, error) {
		return GetRepository(repository, targetOrg, targetToken)
	})

	if err != nil {
		log.Println("failed to migrate repository: ", err)
		return err
	}

	return nil
}

func exponentialBackoff(fn func() (Repository, error)) error {
	for i := 0; i < 10; i++ {
		_, err := fn()
		if err == nil {
			return nil
		} else {
			log.Println("[⏳] retrying in", math.Pow(2, float64(i)), "seconds")
		}
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
	}
	return errors.New("exceeded maximum number of attempts")
}
