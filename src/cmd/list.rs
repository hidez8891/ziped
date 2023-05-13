use encoding_rs::Encoding;
use std::collections::VecDeque;
use std::error::Error;
use std::fs::File;
use std::process::exit;
use wildmatch::WildMatch;
use zip;

use super::Command;
use crate::cli::GlobalOption;

pub(crate) struct List {
    filter: String,
}

impl List {
    pub(crate) fn parse(args: &mut VecDeque<String>) -> Command {
        let mut options = List {
            filter: String::from("*"),
        };

        while let Some(arg) = args.pop_front() {
            match arg.as_str() {
                "-h" | "--help" => {
                    Self::usage();
                }
                opt if opt.starts_with("--filter=") => {
                    options.filter = String::from(opt.split("=").last().unwrap());
                }
                _ => {
                    args.push_front(arg);
                    break;
                }
            }
        }

        Command::List(options)
    }

    #[rustfmt::skip]
    pub(crate) fn usage() -> ! {
        println!(
r#"Usage: ziped ls [OPTION] <PATH>...

Options:
      --filter=<pattern>  Patterns to filter filenames [default=*]
  -h, --help              Print this message

Arguments:
  PATH  Path to zip archive(s)
"#
        );
        exit(0)
    }

    pub(crate) fn run(&self, opt: &GlobalOption, path: &str) -> Result<(), Box<dyn Error>> {
        let reader = File::open(path)?;
        let mut zip = zip::ZipArchive::new(reader)?;

        let path_encoder = Encoding::for_label(opt.path_encoding.as_bytes())
            .ok_or(format!("Invalid path encoding '{}'", opt.path_encoding))?;

        let matcher = WildMatch::new(&self.filter);
        for i in 0..zip.len() {
            let file = zip.by_index(i)?;

            let (name, _, _) = path_encoder.decode(file.name_raw());
            let name = name.into_owned();
            if matcher.matches(name.as_str()) {
                println!("{}", name);
            }
        }

        Ok(())
    }
}
