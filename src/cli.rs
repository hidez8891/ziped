use std::{collections::VecDeque, process::exit};

use crate::cmd::{Command, List, Remove};

pub(crate) enum OutputOption {
    None,
    Path(String),
    Dir(String),
    Overwrite,
}

pub(crate) struct GlobalOption {
    pub(crate) path_encoding: String,
    pub(crate) output: OutputOption,
}

impl Default for GlobalOption {
    fn default() -> Self {
        GlobalOption {
            path_encoding: "ascii".to_owned(),
            output: OutputOption::None,
        }
    }
}

pub(crate) struct Cli {
    pub(crate) option: GlobalOption,
    pub(crate) commands: Vec<Command>,
    pub(crate) paths: Vec<String>,
}

impl Cli {
    #[rustfmt::skip]
    pub(crate) fn usage() -> ! {
        println!(
r#"Usage: ziped [OPTION] <COMMAND> <PATH>...

Options:
  -e, --encoding=<encoding>  Encoding of file path (default: utf-8)
  -h, --help                 Print this message
  -o, --output=<path>        Output archive path
      --output-dir=<path>    Output directory path
      --overwrite            Overwrite existing files
  -v, --version              Print version

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
    opt: GlobalOption,
    args: VecDeque<String>,
}

impl Parser {
    pub(crate) fn parse(args: Vec<String>) -> Cli {
        if args.len() < 2 {
            Cli::usage();
        }

        let mut parser = Parser {
            opt: GlobalOption::default(),
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
                    commands.push(List::parse(&mut self.args));
                }
                "rm" => {
                    // subcommand: ls
                    commands.push(Remove::parse(&mut self.args));
                }
                opt if opt.starts_with("-") => {
                    // global option
                    self.parse_global_option(arg);
                }
                path if path.ends_with(".zip") => {
                    // positional argument
                    self.args.push_front(arg);
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

        Cli {
            option: std::mem::take(&mut self.opt),
            commands,
            paths,
        }
    }

    fn parse_global_option(&mut self, arg: String) {
        match arg.as_str() {
            "-h" | "--help" => {
                Cli::usage();
            }
            "-v" | "--version" => {
                Cli::version();
            }
            opt if opt.starts_with("-e=") || opt.starts_with("--encoding=") => {
                self.opt.path_encoding = String::from(opt.split("=").last().unwrap());
            }
            opt if opt.starts_with("-o=") || opt.starts_with("--output=") => {
                self.opt.output = OutputOption::Path(String::from(opt.split("=").last().unwrap()));
            }
            opt if opt.starts_with("--output-dir=") => {
                self.opt.output = OutputOption::Dir(String::from(opt.split("=").last().unwrap()));
            }
            "--overwrite" => {
                self.opt.output = OutputOption::Overwrite;
            }
            _ => {
                eprintln!("Unknown option: {}", arg);
                exit(1);
            }
        }
    }
}
