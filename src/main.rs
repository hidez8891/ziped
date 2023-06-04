use cli::OutputWriter;
use glob::glob;
use std::error::Error;
use std::{env, fs};

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

    let paths = expand_wildcard_path(&cli.paths).expect("Failed to read path");
    for filepath in paths.iter() {
        let reader = fs::File::open(filepath).expect("Failed to open file for read");
        let mut writer =
            OutputWriter::new(&cli.option.output, filepath).expect("Failed to open file for write");

        match &cli.commands[0] {
            cmd::Command::List(cmd) => {
                cmd.run(&cli.option, &reader)
                    .expect("Failed to read zip-archive");
            }
            cmd::Command::Remove(cmd) => {
                cmd.run(&cli.option, &reader, &writer)
                    .expect("Failed to access zip-archive");
            }
        }

        writer.close().expect("Failed to close file for write");
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
