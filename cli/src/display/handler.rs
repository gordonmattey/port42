use crate::display::OutputFormat;
use anyhow::Result;

/// Base trait for command handlers that support different output formats
pub trait CommandHandler {
    type Output;
    
    /// Execute the command and return the raw output
    fn execute(&mut self) -> Result<Self::Output>;
    
    /// Display the output according to the format
    fn display(&self, output: &Self::Output, format: OutputFormat) -> Result<()>;
    
    /// Run the command with the specified output format
    fn run_with_format(&mut self, format: OutputFormat) -> Result<()> {
        let output = self.execute()?;
        self.display(&output, format)?;
        Ok(())
    }
}

/// Builder pattern for handlers
pub struct HandlerBuilder<T> {
    handler: T,
    output_format: OutputFormat,
}

impl<T> HandlerBuilder<T> {
    pub fn new(handler: T) -> Self {
        Self {
            handler,
            output_format: OutputFormat::Plain,
        }
    }
    
    pub fn with_output_format(mut self, format: OutputFormat) -> Self {
        self.output_format = format;
        self
    }
    
    pub fn build(self) -> (T, OutputFormat) {
        (self.handler, self.output_format)
    }
}