use clap::Args;

#[derive(Args, Debug)]
pub struct VerifyArgs {
    #[arg(short, long, default_value = "all")]
    pub scope: String,
}

fn check_python() -> bool {
    true
}

fn check_java() -> bool {
    false
}

fn check_cobol() -> bool {
    false
}

fn check_nodejs() -> bool {
    true
}

fn check_all() -> bool {
    check_cobol() && check_java() && check_nodejs() && check_python()
}

pub fn run(args: VerifyArgs) {

    let result = match args.scope.as_str() {
        "java" => check_java(),
        "python" => check_python(),
        "cobol" => check_cobol(),
        "nodejs" => check_nodejs(),
        "all" => check_all(),
        _other => check_all(),
    };

    if result {
        println!("Verification Success")
    } else {
        println!("Verification Failed")
    };
}