use glob::glob;
use std::env;
use std::error::Error;

mod cli;
mod cmd;

fn main() {
    let args = env::args().collect::<Vec<String>>();
    let cli = cli::Parser::parse(args);

    if cli.commands.is_empty() || cli.paths.is_empty() {
        println!("{}", cli.commands.len());
        println!("{}", cli.paths.len());
        cli::Cli::usage();
    }

    match &cli.commands[0] {
        cmd::Command::List(cmd) => {
            let paths = expand_wildcard_path(&cli.paths).expect("Failed to read path");
            for filepath in paths.iter() {
                cmd.run(filepath).expect("Failed to read zip-archive");
            }
        }
    }
}

fn expand_wildcard_path(path: &Vec<String>) -> Result<Vec<String>, Box<dyn Error>> {
    let mut expand = Vec::new();

    for pattern in path {
        if pattern.contains("*") {
            for entry in glob(pattern)? {
                let path = entry?;
                expand.push(path.to_str().unwrap().to_string());
            }
        } else {
            expand.push(pattern.clone());
        }
    }

    Ok(expand)
}
