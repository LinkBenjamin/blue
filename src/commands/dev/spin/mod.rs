pub mod aws_lambda;
pub mod python_fastapi; // 1. Declare module

use aws_lambda::AwsLambdaArgs;
use clap::{Args, Subcommand};
use python_fastapi::PythonFastApiArgs; // 2. Import args

#[derive(Args, Debug)]
pub struct SpinArgs {
    #[command(subcommand)]
    pub template: SpinTemplates,
}

#[derive(Subcommand, Debug)]
pub enum SpinTemplates {
    /// Spin up an AWS Lambda environment template
    AwsLambda(AwsLambdaArgs),

    /// Spin up a Python FastAPI environment template
    #[command(name = "python-fastapi")]
    PythonFastApi(PythonFastApiArgs),
}

pub fn run(args: SpinArgs) {
    match args.template {
        SpinTemplates::AwsLambda(lambda_args) => aws_lambda::run(lambda_args),
        SpinTemplates::PythonFastApi(fastapi_args) => python_fastapi::run(fastapi_args), // 4. Dispatch handler
    }
}