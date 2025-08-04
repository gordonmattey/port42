use anyhow::Result;

#[derive(Debug, Clone, Copy)]
pub enum OutputFormat {
    Plain,
    Json,
    Table,
}

pub trait Displayable {
    fn display(&self, format: OutputFormat) -> Result<()>;
}

// Re-export components
pub mod components;
pub use components::*;