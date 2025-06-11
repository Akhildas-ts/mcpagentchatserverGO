package git

import (
	"fmt"
	"os/exec"
)

// CloneRepository clones a Git repository to the specified directory
func CloneRepository(repoURL, targetDir, branch string) error {
	// Clone repository
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Checkout specific branch if specified
	if branch != "" && branch != "main" && branch != "master" {
		cmd = exec.Command("git", "checkout", branch)
		cmd.Dir = targetDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
		}
		fmt.Printf("Checked out branch: %s\n", branch)
	}

	return nil
}