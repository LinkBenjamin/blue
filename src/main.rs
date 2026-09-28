mod cli;
mod commands;

use clap::Parser;
use cli::{Cli, Commands};

fn main() {
    let args = Cli::parse();

    match args.command {
        Commands::Dev(dev_args) => commands::dev::run(dev_args),
    }
}