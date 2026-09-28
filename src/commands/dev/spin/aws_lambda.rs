use clap::Args;

#[derive(Args, Debug)]
pub struct AwsLambdaArgs {
    /// AWS Account ID
    #[arg(long)]
    pub account_id: Option<String>,

    /// Catch-all for extra arguments
    #[arg(allow_hyphen_values = true, trailing_var_arg = true)]
    pub extra_args: Vec<String>,
}

pub fn run(args: AwsLambdaArgs) {
    if let Some(account_id) = args.account_id {
        println!("Spinning up AWS Lambda template for account: {}", account_id);
    } else {
        println!("Spinning up AWS Lambda template...");
    }

    if !args.extra_args.is_empty() {
        println!("Additional flags: {:?}", args.extra_args);
    }
}