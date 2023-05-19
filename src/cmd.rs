mod list;
mod remove;

pub(crate) use list::List;
pub(crate) use remove::Remove;

pub(crate) enum Command {
    List(List),
    Remove(Remove),
}
