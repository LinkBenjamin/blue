pub mod verify;

use verify::VerifyArgs;
use clap::{Args, Subcommand};

#[derive(Args, Debug)]
pub struct EnvArgs {
    #[command(subcommand)]
    pub template: EnvTemplates,
}

#[derive(Subcommand, Debug)]
pub enum EnvTemplates {
    /// Spin up a Verify environment template
    Verify(VerifyArgs),
}

pub fn run(args: EnvArgs) {
    match args.template {
        EnvTemplates::Verify(verify_args) => verify::run(verify_args),
    }
}