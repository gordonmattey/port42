use colored::*;
use prettytable::{Table, Row, Cell, format};

pub struct TableBuilder {
    table: Table,
}

impl TableBuilder {
    pub fn new() -> Self {
        let mut table = Table::new();
        // Set a nice format for the table
        table.set_format(*format::consts::FORMAT_NO_LINESEP_WITH_TITLE);
        Self { table }
    }
    
    pub fn add_header(&mut self, headers: Vec<&str>) -> &mut Self {
        let cells: Vec<Cell> = headers.iter()
            .map(|h| Cell::new(h).style_spec("Fb"))
            .collect();
        self.table.set_titles(Row::new(cells));
        self
    }
    
    pub fn add_row(&mut self, values: Vec<String>) -> &mut Self {
        let cells: Vec<Cell> = values.iter()
            .map(|v| Cell::new(v))
            .collect();
        self.table.add_row(Row::new(cells));
        self
    }
    
    pub fn print(&self) {
        self.table.printstd();
    }
}

pub fn format_timestamp_relative(timestamp: u64) -> String {
    use std::time::{SystemTime, UNIX_EPOCH, Duration};
    
    let timestamp_secs = timestamp / 1000;
    let past = UNIX_EPOCH + Duration::from_secs(timestamp_secs);
    
    match SystemTime::now().duration_since(past) {
        Ok(duration) => {
            let secs = duration.as_secs();
            match secs {
                0..=59 => format!("{} seconds ago", secs),
                60..=3599 => format!("{} minutes ago", secs / 60),
                3600..=86399 => format!("{} hours ago", secs / 3600),
                _ => format!("{} days ago", secs / 86400),
            }
        }
        Err(_) => "in the future".to_string(),
    }
}

// Helper for consistent status indicators
pub struct StatusIndicator;

impl StatusIndicator {
    pub fn success() -> ColoredString {
        "✅".green()
    }
    
    pub fn error() -> ColoredString {
        "❌".red()
    }
}

// Progress indicator for long operations
pub struct ProgressIndicator {
    message: String,
    spinner_chars: Vec<char>,
    current: usize,
}

impl ProgressIndicator {
    pub fn new(message: &str) -> Self {
        Self {
            message: message.to_string(),
            spinner_chars: vec!['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'],
            current: 0,
        }
    }
    
    pub fn tick(&mut self) {
        print!("\r{} {} ", 
            self.spinner_chars[self.current].to_string().cyan(),
            self.message
        );
        use std::io::{self, Write};
        io::stdout().flush().unwrap();
        
        self.current = (self.current + 1) % self.spinner_chars.len();
    }
    
    pub fn finish(&self, message: &str) {
        println!("\r{} {}", StatusIndicator::success(), message);
    }
}