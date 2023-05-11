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
r#"Usage: ziped [OPTION] <COMMAND> <PATH>...

Options:
  -h, --help     Print this message
  -v, --version  Print version

Commands:
  ls    List files in zip archive

Arguments:
  PATH  Path to zip archive(s)
"#
        );
        exit(0)
    }

    pub fn version() -> ! {
        let version = env!("CARGO_PKG_VERSION");
        println!("ziped v{}", version);
        exit(0)
    }

    fn parse_(&mut self) -> Cli {
        let mut commands = Vec::new();

        while self.index < self.args.len() {
            let arg = &self.args[self.index];

            match arg.as_str() {
                "ls" => {
                    // subcommand: ls
                    commands.push(self.parse_ls());
                }
                opt if opt.starts_with("-") => {
                    // global option
                    self.parse_global_option();
                }
                path if path.ends_with(".zip") => {
                    // positional argument
                    break;
                }
                _ => {
                    // unknown subcommand
                    eprintln!("Unknown subcommand: {}", arg);
                    exit(1);
                }
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

    fn parse_global_option(&mut self) -> ! {
        let arg = &self.args[self.index];

        match arg.as_str() {
            "-h" | "--help" => {
                Self::usage();
            }
            "-v" | "--version" => {
                Self::version();
            }
            _ => {
                eprintln!("Unknown option: {}", arg);
                exit(1);
            }
        }
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

fn exec_list(path: &String) -> Result<(), Box<dyn Error>> {
    let reader = File::open(path)?;
    let mut zip = zip::ZipArchive::new(reader)?;

    for i in 0..zip.len() {
        let file = zip.by_index(i)?;
        println!("{}", file.name());
    }

    Ok(())
}
