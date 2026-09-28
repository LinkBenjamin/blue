use clap::Args;
use std::fs;
use std::path::Path;
use std::process::Command;

#[derive(Args, Debug)]
pub struct PythonFastApiArgs {
    /// Name of the project / subfolder to create
    #[arg(short, long, default_value = "myproject")]
    pub name: String,

    /// Port number for local development server
    #[arg(short, long, default_value_t = 8000)]
    pub port: u16,
}

pub fn run(args: PythonFastApiArgs) {
    let project_dir = Path::new(&args.name);

    if project_dir.exists() {
        eprintln!("Error: Directory '{}' already exists.", args.name);
        return;
    }

    println!("Spinning up Python FastAPI project in './{}'...", args.name);

    if let Err(e) = fs::create_dir_all(project_dir) {
        eprintln!("Failed to create project directory: {}", e);
        return;
    }

    create_starter_files(project_dir, &args.name, args.port);
    init_git_repo(project_dir);
    create_virtualenv(project_dir);

    print_next_steps(&args.name);
}

fn create_starter_files(project_dir: &Path, project_name: &str, port: u16) {
    let main_py = format!(
        r#"'''Project entry point for FastAPI Server'''
from fastapi import FastAPI
import uvicorn

app = FastAPI()

@app.get("/")
def read_root():
    '''Hello World Method, replace as appropriate'''
    return {{"message": "Hello World"}}

if __name__ == "__main__":
    uvicorn.run("main:app", host="127.0.0.1", port={}, reload=True)
"#,
        port
    );

    // Standard PEP 621 pyproject.toml format
    let pyproject_toml = format!(
        r#"[project]
name = "{}"
version = "0.1.0"
description = "FastAPI application generated with blue"
readme = "README.md"
requires-python = ">=3.12"
dependencies = [
    "fastapi",
    "uvicorn[standard]",
]

[build-system]
requires = ["setuptools>=61.0"]
build-backend = "setuptools.build_meta"
"#,
        project_name
    );

    let gitignore = "__pycache__/\n*.pyc\n.venv/\n.env\n";
    let readme = format!("# {}\n\nFastAPI project created with `blink`.", project_name);

    fs::write(project_dir.join("main.py"), main_py).expect("Failed to create main.py");
    fs::write(project_dir.join("pyproject.toml"), pyproject_toml).expect("Failed to create pyproject.toml");
    fs::write(project_dir.join("README.md"), readme).expect("Failed to create README.md");
    fs::write(project_dir.join(".gitignore"), gitignore).expect("Failed to create .gitignore");
}

fn init_git_repo(project_dir: &Path) {
    println!("Initializing Git repository...");

    let status = Command::new("git")
        .arg("init")
        .current_dir(project_dir)
        .status();

    match status {
        Ok(s) if s.success() => println!("Git repository initialized successfully."),
        Ok(s) => eprintln!("'git init' failed with exit status: {}", s),
        Err(e) => eprintln!("Failed to execute 'git init': {}. Make sure Git is installed.", e),
    }
}

fn create_virtualenv(project_dir: &Path) {
    println!("Creating Python virtual environment (.venv)...");

    let python_cmd = if Command::new("python3").arg("--version").output().is_ok() {
        "python3"
    } else {
        "python"
    };

    let status = Command::new(python_cmd)
        .args(["-m", "venv", ".venv"])
        .current_dir(project_dir)
        .status();

    match status {
        Ok(s) if s.success() => println!("Virtual environment created at .venv/"),
        Ok(s) => eprintln!("Failed to create venv, exit status: {}", s),
        Err(e) => eprintln!("Failed to run python venv command: {}", e),
    }
}

fn print_next_steps(project_name: &str) {
    println!("\nProject created successfully!\n");
    println!("Before you start your coding, activate your virtual environment:\n");
    println!("  cd {}", project_name);

    if cfg!(windows) {
        println!("  .venv\\Scripts\\activate");
    } else {
        println!("  source .venv/bin/activate");
    }

    // Modern pip supports editable installs directly from pyproject.toml
    println!("  pip install -e .\n");
    println!("Then you can run your project like this:\n");
    println!("  python main.py\n");
}