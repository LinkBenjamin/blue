package spin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	projectName string
	port        int
)

var pythonFastApiCmd = &cobra.Command{
	Use:   "python-fastapi",
	Short: "Spin up a Python FastAPI environment template",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPythonFastApi(projectName, port)
	},
}

func init() {
	SpinCmd.AddCommand(pythonFastApiCmd)

	// Define command flags
	pythonFastApiCmd.Flags().StringVarP(&projectName, "name", "n", "myproject", "Name of the project / subfolder to create")
	pythonFastApiCmd.Flags().IntVarP(&port, "port", "p", 8000, "Port number for local development server")
}

func runPythonFastApi(name string, port int) error {
	projectDir := name

	// 1. Check if target directory already exists
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("error: Directory '%s' already exists", name)
	}

	fmt.Printf("Spinning up Python FastAPI project in './%s'...\n", name)

	// 2. Create directory structure
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// 3. Create basic starter files
	if err := createStarterFiles(projectDir, port); err != nil {
		return err
	}

	// 4. Run `git init` inside the project folder
	if err := initGitRepo(projectDir); err != nil {
		return err
	}

	fmt.Printf("Successfully initialized FastAPI project in './%s'\n", name)
	return nil
}

func createStarterFiles(projectDir string, port int) error {
	mainPy := fmt.Sprintf(`from fastapi import FastAPI
import uvicorn

app = FastAPI()

@app.get("/")
def read_root():
    return {"message": "Hello World"}

if __name__ == "__main__":
    uvicorn.run("main:app", host="127.0.0.1", port=%d, reload=True)
`, port)

	pyprojectToml := `[project]
name = "myproject"
version = "0.1.0"
dependencies = [
    "fastapi",
    "uvicorn[standard]",
]
`
	gitignore := "__pycache__/\n*.pyc\n.venv/\n.env\n"

	files := map[string]string{
		"main:app.py":    mainPy,
		"pyproject.toml": pyprojectToml,
		".gitignore":     gitignore,
	}

	for filename, content := range files {
		path := filepath.Join(projectDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}

func initGitRepo(projectDir string) error {
	fmt.Println("Initializing Git repository...")

	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir // Equivalent to running git init inside projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("'git init' failed: %w", err)
	}

	fmt.Println("Git repository initialized successfully.")
	return nil
}
