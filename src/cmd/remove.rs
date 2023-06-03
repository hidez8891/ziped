use encoding_rs::Encoding;
use std::collections::VecDeque;
use std::error::Error;
use std::fs;
use std::process::exit;
use wildmatch::WildMatch;
use zip;

use super::Command;
use crate::cli::{GlobalOption, OutputWriter};

pub(crate) struct Remove {
    filters: Vec<String>,
}

impl Remove {
    pub(crate) fn parse(args: &mut VecDeque<String>) -> Command {
        let mut options = Remove {
            filters: Vec::new(),
        };

        while let Some(arg) = args.pop_front() {
            match arg.as_str() {
                "-h" | "--help" => {
                    Self::usage();
                }
                opt if opt.starts_with("--filter=") => {
                    options
                        .filters
                        .push(String::from(opt.split("=").last().unwrap()));
                }
                _ => {
                    args.push_front(arg);
                    break;
                }
            }
        }

        Command::Remove(options)
    }

    #[rustfmt::skip]
    pub(crate) fn usage() -> ! {
        println!(
r#"Usage: ziped rm [OPTION] <PATH>...

Options:
      --filter=<pattern>  Patterns to filter filenames
                          This option can be specified multiple times
  -r, --recursive         Recursively remove files in directory
  -h, --help              Print this message

Arguments:
  PATH  Path to zip archive(s)
"#
        );
        exit(0)
    }

    pub(crate) fn run(&self, opt: &GlobalOption, path: &str) -> Result<(), Box<dyn Error>> {
        let reader = fs::File::open(path)?;
        let mut zipr = zip::ZipArchive::new(reader)?;

        let path_encoder = Encoding::for_label(opt.path_encoding.as_bytes())
            .ok_or(format!("Invalid path encoding '{}'", opt.path_encoding))?;

        let matchers: Vec<_> = self
            .filters
            .iter()
            .map(|pattern| WildMatch::new(pattern))
            .collect();

        let mut writer = OutputWriter::new(&opt.output, path)?;
        let mut zipw = zip::ZipWriter::new(&writer);

        for i in 0..zipr.len() {
            let file = zipr.by_index(i)?;

            let filepath = if let Ok(path) = std::str::from_utf8(file.name_raw()) {
                path.to_string()
            } else {
                let (path, _, _) = path_encoder.decode(file.name_raw());
                path.into_owned()
            };

            if matchers
                .iter()
                .any(|matcher| matcher.matches(filepath.as_str()))
            {
                continue;
            }

            // BUG-PATCH: zip-rs executes a garbled file copy.
            let name_raw = file.name_raw().to_owned();
            let name = unsafe { std::str::from_utf8_unchecked(&name_raw) };
            zipw.raw_copy_file_rename(file, name)?;
        }

        drop(zipr);

        zipw.finish()?;
        drop(zipw);

        writer.close()?;

        Ok(())
    }
}
