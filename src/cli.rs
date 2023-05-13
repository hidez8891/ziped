use std::{collections::VecDeque, process::exit};

use crate::cmd::{Command, List};

pub(crate) struct Cli {
    pub(crate) commands: Vec<Command>,
    pub(crate) paths: Vec<String>,
}

impl Cli {
    #[rustfmt::skip]
    pub(crate) fn usage() -> ! {
        println!(
r#"Usage: ziped [OPTION] <COMMAND> <PATH>...

Options:
  -h, --help               Print this message
  -o, --output=<path>      Output archive path
      --output-dir=<path>  Output directory path
      --overwrite          Overwrite existing files
  -v, --version            Print version

Commands
  ls        List files in zip archive

Commands (Edit)
  add       Add files to zip archive
  mv        Move files in zip archive
  rm        Remove files from zip archive

Commands (Convert)
  conv      Convert image file to other format
  convmv    Convert file path to other encoding
  convexec  Convert file with external command

Arguments:
  PATH      Path to zip archive(s)
"#
        );
        exit(0)
    }

    pub(crate) fn version() -> ! {
        let version = env!("CARGO_PKG_VERSION");
        println!("ziped v{}", version);
        exit(0)
    }
}

pub(crate) struct Parser {
    args: VecDeque<String>,
}

impl Parser {
    pub(crate) fn parse(args: Vec<String>) -> Cli {
        if args.len() < 2 {
            Cli::usage();
        }

        let mut parser = Parser {
            args: args.into_iter().collect(),
        };

        parser.args.pop_front(); // skip program name
        parser.parse_()
    }

    fn parse_(&mut self) -> Cli {
        let mut commands = Vec::new();

        while let Some(arg) = self.args.pop_front() {
            match arg.as_str() {
                "ls" => {
                    // subcommand: ls
                    commands.push(List::parse(&self.args));
                }
                opt if opt.starts_with("-") => {
                    // global option
                    self.parse_global_option(arg);
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

        let paths: Vec<String> = std::mem::take(&mut self.args).into_iter().collect();

        // validate positional arguments
        if let Some(path) = paths.iter().find(|path| path.ends_with(".zip") == false) {
            // unknown subcommand or option
            eprintln!("Unknown subcommand or option: {}", path);
            exit(1);
        }

        Cli { commands, paths }
    }

    fn parse_global_option(&mut self, arg: String) -> ! {
        match arg.as_str() {
            "-h" | "--help" => {
                Cli::usage();
            }
            "-v" | "--version" => {
                Cli::version();
            }
            _ => {
                eprintln!("Unknown option: {}", arg);
                exit(1);
            }
        }
    }
}
