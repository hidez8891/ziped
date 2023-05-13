use std::collections::VecDeque;
use std::error::Error;
use std::fs::File;
use std::process::exit;
use zip;

use super::Command;

pub(crate) struct List {}

impl List {
    pub(crate) fn parse(args: &mut VecDeque<String>) -> Command {
        while let Some(arg) = args.pop_front() {
            match arg.as_str() {
                "-h" | "--help" => {
                    Self::usage();
                }
                _ => {
                    args.push_front(arg);
                    break;
                }
            }
        }

        Command::List(List {})
    }

    #[rustfmt::skip]
    pub(crate) fn usage() -> ! {
        println!(
r#"Usage: ziped ls [OPTION] <PATH>...

Options:
  -h, --help  Print this message

Arguments:
  PATH  Path to zip archive(s)
"#
        );
        exit(0)
    }

    pub(crate) fn run(&self, path: &str) -> Result<(), Box<dyn Error>> {
        let reader = File::open(path)?;
        let mut zip = zip::ZipArchive::new(reader)?;

        for i in 0..zip.len() {
            let file = zip.by_index(i)?;
            println!("{}", file.name());
        }

        Ok(())
    }
}
