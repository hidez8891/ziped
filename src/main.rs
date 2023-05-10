use std::fs::File;

use clap::{Parser, Subcommand};
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
            for filepath in path {
                exec_list(filepath).expect("failed");
            }
        }
    }
}

fn exec_list(path: &String) -> zip::result::ZipResult<()> {
    let reader = File::open(path)?;
    let mut zip = zip::ZipArchive::new(reader)?;

    for i in 0..zip.len() {
        let file = zip.by_index(i)?;
        println!("{}", file.name());
    }

    Ok(())
}
