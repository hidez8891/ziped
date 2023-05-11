use clap::{Parser, Subcommand};
use glob::glob;
use std::{error::Error, fs::File};
use zip;

#[derive(Parser)]
#[command(version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// List zip-archive contents
    Ls {
        /// zip-archive path(s)
        path: Vec<String>,
    },
}

fn main() {
    let cli = Cli::parse();

    match &cli.command {
        Commands::Ls { path } => {
            let path = expand_wildcard_path(path).expect("Failed to read path");
            for filepath in path.iter() {
                exec_list(filepath).expect("Failed to read zip-archive");
            }
        }
    }
}

fn expand_wildcard_path(path: &Vec<String>) -> Result<Vec<String>, Box<dyn Error>> {
    let mut expand = Vec::new();

    for pattern in path {
        for entry in glob(pattern)? {
            let path = entry?;
            expand.push(path.to_str().unwrap().to_string());
        }
    }

    Ok(expand)
}

fn exec_list(path: &String) -> Result<(), Box<dyn Error>> {
    let reader = File::open(path)?;
    let mut zip = zip::ZipArchive::new(reader)?;

    for i in 0..zip.len() {
        let file = zip.by_index(i)?;
        println!("{}", file.name());
    }

    Ok(())
}
