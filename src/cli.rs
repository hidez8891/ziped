use std::collections::VecDeque;
use std::error::Error;
use std::fmt;
use std::fs;
use std::io;
use std::path::PathBuf;
use std::process::exit;
use tempfile::NamedTempFile;

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

enum OutputWriterType {
    None,
    File(fs::File),
    Temp(NamedTempFile),
}

pub(crate) struct OutputWriter {
    type_: OutputWriterType,
    path: PathBuf,
}

impl OutputWriter {
    pub(crate) fn new(option: &OutputOption, path: &str) -> Result<Self, Box<dyn Error>> {
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

    pub(crate) fn close(&mut self) -> io::Result<()> {
        // overwirte input file with temporary file.
        if let OutputWriterType::Temp(_) = &self.type_ {
            let mut temp_ = OutputWriterType::None;
            std::mem::swap(&mut temp_, &mut self.type_);

            let OutputWriterType::Temp(temp) = temp_ else { unreachable!(); };
            let _ = temp.persist(self.path.clone())?;
        }
        Ok(())
    }

    fn as_file(&self) -> &std::fs::File {
        let OutputWriterType::File(file) = &self.type_ else { unreachable!()};
        return file;
    }

    fn as_temp(&self) -> &NamedTempFile {
        let OutputWriterType::Temp(file) = &self.type_ else { unreachable!()};
        return file;
    }
}

impl io::Read for OutputWriter {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read(buf),
            OutputWriterType::Temp(_) => self.as_temp().read(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_vectored(&mut self, bufs: &mut [io::IoSliceMut<'_>]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_vectored(bufs),
            OutputWriterType::Temp(_) => self.as_temp().read_vectored(bufs),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_to_end(&mut self, buf: &mut Vec<u8>) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_to_end(buf),
            OutputWriterType::Temp(_) => self.as_temp().read_to_end(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_to_string(&mut self, buf: &mut String) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_to_string(buf),
            OutputWriterType::Temp(_) => self.as_temp().read_to_string(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_exact(&mut self, buf: &mut [u8]) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_exact(buf),
            OutputWriterType::Temp(_) => self.as_temp().read_exact(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Read for &OutputWriter {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read(buf),
            OutputWriterType::Temp(_) => self.as_temp().read(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_vectored(&mut self, bufs: &mut [io::IoSliceMut<'_>]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_vectored(bufs),
            OutputWriterType::Temp(_) => self.as_temp().read_vectored(bufs),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_to_end(&mut self, buf: &mut Vec<u8>) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_to_end(buf),
            OutputWriterType::Temp(_) => self.as_temp().read_to_end(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_to_string(&mut self, buf: &mut String) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_to_string(buf),
            OutputWriterType::Temp(_) => self.as_temp().read_to_string(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn read_exact(&mut self, buf: &mut [u8]) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().read_exact(buf),
            OutputWriterType::Temp(_) => self.as_temp().read_exact(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Write for OutputWriter {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write(buf),
            OutputWriterType::Temp(_) => self.as_temp().write(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn flush(&mut self) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().flush(),
            OutputWriterType::Temp(_) => self.as_temp().flush(),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_vectored(&mut self, bufs: &[io::IoSlice<'_>]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write_vectored(bufs),
            OutputWriterType::Temp(_) => self.as_temp().write_vectored(bufs),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_all(&mut self, buf: &[u8]) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write_all(buf),
            OutputWriterType::Temp(_) => self.as_temp().write_all(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_fmt(&mut self, fmt: fmt::Arguments<'_>) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write_fmt(fmt),
            OutputWriterType::Temp(_) => self.as_temp().write_fmt(fmt),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Write for &OutputWriter {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write(buf),
            OutputWriterType::Temp(_) => self.as_temp().write(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn flush(&mut self) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().flush(),
            OutputWriterType::Temp(_) => self.as_temp().flush(),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_vectored(&mut self, bufs: &[io::IoSlice<'_>]) -> io::Result<usize> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write_vectored(bufs),
            OutputWriterType::Temp(_) => self.as_temp().write_vectored(bufs),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_all(&mut self, buf: &[u8]) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write_all(buf),
            OutputWriterType::Temp(_) => self.as_temp().write_all(buf),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }

    fn write_fmt(&mut self, fmt: fmt::Arguments<'_>) -> io::Result<()> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().write_fmt(fmt),
            OutputWriterType::Temp(_) => self.as_temp().write_fmt(fmt),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Seek for OutputWriter {
    fn seek(&mut self, pos: io::SeekFrom) -> io::Result<u64> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().seek(pos),
            OutputWriterType::Temp(_) => self.as_temp().seek(pos),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}

impl io::Seek for &OutputWriter {
    fn seek(&mut self, pos: io::SeekFrom) -> io::Result<u64> {
        match &self.type_ {
            OutputWriterType::File(_) => self.as_file().seek(pos),
            OutputWriterType::Temp(_) => self.as_temp().seek(pos),
            OutputWriterType::None => Err(io::Error::new(io::ErrorKind::Other, "writer is closed")),
        }
    }
}
