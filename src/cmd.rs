mod list;

pub(crate) use list::List;

pub(crate) enum Command {
    List(List),
}
