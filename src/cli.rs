use clap::{Parser, Subcommand};
use crate::commands::dev::DevArgs;

#[derive(Parser)]
#[command(name = "blink", version, about = "A modular Rust CLI tool")]
pub struct Cli {
    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand)]
pub enum Commands {
    /// Development utility tools
    Dev(DevArgs),
}