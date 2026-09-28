pub mod spin;
pub mod environment;

use clap::{Args, Subcommand};
use spin::SpinArgs;
use environment::EnvArgs;

#[derive(Args, Debug)]
pub struct DevArgs {
    #[command(subcommand)]
    pub command: DevCommands,
}

#[derive(Subcommand, Debug)]
pub enum DevCommands {
    /// Spin up a development environment/template
    Spin(SpinArgs),
    Env(EnvArgs),
}

pub fn run(args: DevArgs) {
    match args.command {
        DevCommands::Spin(spin_args) => spin::run(spin_args),
        DevCommands::Env(env_args) => environment::run(env_args),
    }
}