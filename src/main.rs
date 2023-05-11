use glob::glob;
use std::{env, error::Error, fs::File, process::exit};
use zip;

struct Cli {
    commands: Vec<Commands>,
    paths: Vec<String>,
}

struct CliParser {
    args: Vec<String>,
    index: usize,
}

impl CliParser {
    pub fn parse() -> Cli {
        let args = env::args().collect::<Vec<String>>();

        if args.len() < 2 {
            Self::usage();
        }

        let mut parser = CliParser {
            args,
            index: 1, // skip program name
        };

        parser.parse_()
    }

    #[rustfmt::skip]
    pub fn usage() -> ! {
        println!(
r#"Usage: ziped <COMMAND> <PATH>...

Commands:
  ls     List files in zip archive
  help   Print this message

Arguments:
  PATH   Path to zip archive(s)
"#
        );
        exit(0)
    }

    fn parse_(&mut self) -> Cli {
        let mut commands = Vec::new();

        while self.index < self.args.len() {
            let arg = &self.args[self.index];

            if arg == "ls" {
                // subcommand: ls
                commands.push(self.parse_ls());
            } else if arg.starts_with("-") {
                // global option
                eprintln!("Unknown option: {}", arg);
                exit(1);
            } else if arg.ends_with(".zip") {
                // positional argument
                break;
            } else {
                // unknown subcommand
                eprintln!("Unknown subcommand: {}", arg);
                exit(1);
            }
        }

        let paths = self.args[self.index..].to_vec();

        // validate positional arguments
        if let Some(path) = paths.iter().find(|path| path.ends_with(".zip") == false) {
            // unknown subcommand or option
            eprintln!("Unknown subcommand or option: {}", path);
            exit(1);
        }

        Cli { commands, paths }
    }

    fn parse_ls(&mut self) -> Commands {
        self.index += 1; // skip command name

        Commands::List
    }
}

enum Commands {
    List,
}

fn main() {
    let cli = CliParser::parse();

    if cli.commands.is_empty() || cli.paths.is_empty() {
        CliParser::usage();
    }

    match &cli.commands[0] {
        Commands::List => {
            let paths = expand_wildcard_path(&cli.paths).expect("Failed to read path");
            for filepath in paths.iter() {
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
