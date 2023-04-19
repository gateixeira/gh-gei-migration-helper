package cmd

import (
	"fmt"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	fmt.Print(
		"\n\n========================================\nRepository " + *repository.Name + "\n========================================\n")

	// // if *repository.Visibility != "public" {
	// 	fmt.Println("[🔄] Deactivating GHAS settings at source repository")
	// 	github.ChangeGhasRepoSettings(sourceOrg, repository, false, sourceToken)
	// 	fmt.Println("[✅] Done")
	// // }

	// fmt.Println("[🔄] Migrating repository")
	// github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	// fmt.Println("[✅] Done")

	// fmt.Println("[🔄] Deleting branch protections at target")
	// github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	// fmt.Println("[✅] Done")

	// //check if repository is not private
	// if !*repository.Private {
	// 	fmt.Println("[🔄] Repository not private at source. Changing visibility to internal at target")
	// 	github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
	// 	fmt.Println("[✅] Done")
	// }

	fmt.Println("[🔄] Activating GHAS settings at target")
	github.ChangeGhasRepoSettings(targetOrg, repository, true, targetToken)
	fmt.Println("[✅] Finished.")
}
