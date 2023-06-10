use encoding_rs::Encoding;
use std::collections::VecDeque;
use std::error::Error;
use std::io::{Read, Seek, Write};
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

    pub(crate) fn run<R, W>(
        &self,
        opt: &GlobalOption,
        reader: R,
        writer: &mut W,
    ) -> Result<(), Box<dyn Error>>
    where
        R: Read + Seek,
        W: Write,
    {
        let mut zip = zip::ZipArchive::new(reader)?;

        let path_encoder = Encoding::for_label(opt.path_encoding.as_bytes())
            .ok_or(format!("Invalid path encoding '{}'", opt.path_encoding))?;

        let matcher = WildMatch::new(&self.filter);
        for i in 0..zip.len() {
            let file = zip.by_index(i)?;

            let filepath = if let Ok(path) = std::str::from_utf8(file.name_raw()) {
                path.to_string()
            } else {
                let (path, _, _) = path_encoder.decode(file.name_raw());
                path.into_owned()
            };

            if matcher.matches(filepath.as_str()) {
                writeln!(writer, "{}", filepath)?;
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::*;
    use std::io;

    fn setup_reader() -> io::Cursor<Vec<u8>> {
        let mut buf = io::Cursor::new(Vec::new());
        let mut zip = zip::ZipWriter::new(&mut buf);

        let options =
            zip::write::FileOptions::default().compression_method(zip::CompressionMethod::Stored);
        zip.start_file("test01.txt", options).unwrap();
        zip.write(b"test file 01 text").unwrap();
        zip.start_file("test02.txt", options).unwrap();
        zip.write(b"test file 02 text").unwrap();
        zip.start_file("test03.txt", options).unwrap();
        zip.write(b"test file 03 text").unwrap();
        zip.finish().unwrap();
        drop(zip);

        return buf;
    }

    #[test]
    fn show_all() {
        let reader = setup_reader();
        let mut writer = io::Cursor::new(Vec::new());

        let cmd = List {
            filter: String::from("*"),
        };
        let opt = GlobalOption {
            path_encoding: "sjis".to_owned(),
            output: cli::OutputOption::None,
        };

        cmd.run(&opt, reader, &mut writer).unwrap();

        assert_eq!(
            "test01.txt\ntest02.txt\ntest03.txt\n",
            String::from_utf8(writer.into_inner()).unwrap()
        );
    }

    #[test]
    fn target_filter() {
        let reader = setup_reader();
        let mut writer = io::Cursor::new(Vec::new());

        let cmd = List {
            filter: String::from("*2.txt"),
        };
        let opt = GlobalOption {
            path_encoding: "sjis".to_owned(),
            output: cli::OutputOption::None,
        };

        cmd.run(&opt, reader, &mut writer).unwrap();

        assert_eq!(
            "test02.txt\n",
            String::from_utf8(writer.into_inner()).unwrap()
        );
    }
}
