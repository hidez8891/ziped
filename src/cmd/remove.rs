use encoding_rs::Encoding;
use std::collections::VecDeque;
use std::error::Error;
use std::fmt;
use std::fs;
use std::io;
use std::path::PathBuf;
use std::process::exit;
use tempfile::NamedTempFile;
use wildmatch::WildMatch;
use zip;

use super::Command;
use crate::cli::{GlobalOption, OutputOption};

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

        let writer = OutputWriter::new(&opt.output, path)?;
        let mut zipw = zip::ZipWriter::new(writer);

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
        Ok(())
    }
}

enum OutputWriterType {
    None,
    File(fs::File),
    Temp(NamedTempFile),
}

struct OutputWriter {
    type_: OutputWriterType,
    path: PathBuf,
}

impl OutputWriter {
    fn new(option: &OutputOption, path: &str) -> Result<Self, Box<dyn Error>> {
        match option {
            OutputOption::None => {
                return Err(format!("Output is not specified").into());
            }
            OutputOption::Path(path) => {
                let output_path = PathBuf::from(path);
                if output_path.exists() {
                    return Err(format!("'{}' already exists", path).into());
                }

                Ok(OutputWriter {
                    type_: OutputWriterType::File(fs::File::create(output_path.clone())?),
                    path: output_path,
                })
            }
            OutputOption::Dir(dir) => {
                let path_dir = PathBuf::from(dir);
                if path_dir.exists() {
                    if !path_dir.is_dir() {
                        return Err(format!("'{}' is not a directory", dir).into());
                    }
                } else {
                    fs::create_dir_all(path_dir.clone())?;
                }

                let path_input = PathBuf::from(path);
                let filename = path_input.file_name().unwrap();

                let output_path = path_dir.join(filename);
                Ok(OutputWriter {
                    type_: OutputWriterType::File(fs::File::create(output_path.clone())?),
                    path: output_path,
                })
            }
            OutputOption::Overwrite => {
                let output_path = PathBuf::from(path);
                let path_dir = output_path.parent().unwrap();

                Ok(OutputWriter {
                    type_: OutputWriterType::Temp(NamedTempFile::new_in(path_dir)?),
                    path: output_path,
                })
            }
        }
    }

    fn close(&mut self) -> io::Result<()> {
        if let OutputWriterType::Temp(_) = &self.type_ {
            let mut temp_ = OutputWriterType::None;
            std::mem::swap(&mut temp_, &mut self.type_);

            let OutputWriterType::Temp(temp) = temp_ else { unreachable!(); };
            println!("{:?}", &temp);
            let _ = temp.persist(self.path.clone())?;
        }
        Ok(())
    }
}

impl Drop for OutputWriter {
    fn drop(&mut self) {
        if let Err(err) = self.close() {
            eprintln!("OutputWriter drop failed: {:?}", err);
        }
    }
}

impl io::Read for OutputWriter {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.read(buf),
            OutputWriterType::Temp(temp) => temp.read(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_vectored(&mut self, bufs: &mut [io::IoSliceMut<'_>]) -> io::Result<usize> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.read_vectored(bufs),
            OutputWriterType::Temp(temp) => temp.read_vectored(bufs),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_to_end(&mut self, buf: &mut Vec<u8>) -> io::Result<usize> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.read_to_end(buf),
            OutputWriterType::Temp(temp) => temp.read_to_end(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_to_string(&mut self, buf: &mut String) -> io::Result<usize> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.read_to_string(buf),
            OutputWriterType::Temp(temp) => temp.read_to_string(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_exact(&mut self, buf: &mut [u8]) -> io::Result<()> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.read_exact(buf),
            OutputWriterType::Temp(temp) => temp.read_exact(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Write for OutputWriter {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.write(buf),
            OutputWriterType::Temp(temp) => temp.write(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn flush(&mut self) -> io::Result<()> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.flush(),
            OutputWriterType::Temp(temp) => temp.flush(),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_vectored(&mut self, bufs: &[io::IoSlice<'_>]) -> io::Result<usize> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.write_vectored(bufs),
            OutputWriterType::Temp(temp) => temp.write_vectored(bufs),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_all(&mut self, buf: &[u8]) -> io::Result<()> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.write_all(buf),
            OutputWriterType::Temp(temp) => temp.write_all(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_fmt(&mut self, fmt: fmt::Arguments<'_>) -> io::Result<()> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.write_fmt(fmt),
            OutputWriterType::Temp(temp) => temp.write_fmt(fmt),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Seek for OutputWriter {
    fn seek(&mut self, pos: io::SeekFrom) -> io::Result<u64> {
        match &mut self.type_ {
            OutputWriterType::File(file) => file.seek(pos),
            OutputWriterType::Temp(temp) => temp.seek(pos),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}
