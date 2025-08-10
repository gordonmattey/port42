//! Animated spinner for loading states
//! 
//! Provides a simple spinner animation to show while waiting for AI responses

use std::io::{self, Write};
use std::sync::mpsc::{self, Sender};
use std::thread;
use std::time::Duration;
use colored::*;
use crossterm::{cursor, execute};

pub struct Spinner {
    handle: Option<thread::JoinHandle<()>>,
    stop_sender: Option<Sender<()>>,
}

impl Spinner {
    pub fn new(message: &str) -> io::Result<Self> {
        let (tx, rx) = mpsc::channel();
        let msg = message.to_string();
        
        let handle = thread::spawn(move || {
            let frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
            let mut frame_idx = 0;
            
            // Hide cursor
            let _ = execute!(io::stdout(), cursor::Hide);
            
            loop {
                // Check if we should stop
                if rx.try_recv().is_ok() {
                    break;
                }
                
                // Print spinner frame
                print!("\r{} {}  ", 
                       frames[frame_idx].bright_cyan(), 
                       msg.dimmed());
                let _ = io::stdout().flush();
                
                frame_idx = (frame_idx + 1) % frames.len();
                
                // Sleep for animation speed
                thread::sleep(Duration::from_millis(100));
            }
            
            // Clear the line and show cursor again
            print!("\r{}", " ".repeat(msg.len() + 10));
            print!("\r");
            let _ = execute!(io::stdout(), cursor::Show);
            let _ = io::stdout().flush();
        });
        
        Ok(Spinner {
            handle: Some(handle),
            stop_sender: Some(tx),
        })
    }
    
    pub fn stop(mut self) {
        if let Some(sender) = self.stop_sender.take() {
            let _ = sender.send(());
        }
        
        if let Some(handle) = self.handle.take() {
            let _ = handle.join();
        }
    }
    
    pub fn with_message(message: &str) -> SpinnerGuard {
        SpinnerGuard::new(message)
    }
}

impl Drop for Spinner {
    fn drop(&mut self) {
        if let Some(sender) = self.stop_sender.take() {
            let _ = sender.send(());
        }
        
        if let Some(handle) = self.handle.take() {
            let _ = handle.join();
        }
    }
}

/// RAII guard for spinner that automatically stops when dropped
pub struct SpinnerGuard {
    spinner: Option<Spinner>,
}

impl SpinnerGuard {
    pub fn new(message: &str) -> Self {
        let spinner = Spinner::new(message).ok();
        SpinnerGuard { spinner }
    }
    
    pub fn stop(mut self) {
        if let Some(spinner) = self.spinner.take() {
            spinner.stop();
        }
    }
}

impl Drop for SpinnerGuard {
    fn drop(&mut self) {
        if let Some(spinner) = self.spinner.take() {
            spinner.stop();
        }
    }
}

/// Simple inline spinner without threads for simpler cases
pub struct SimpleSpinner {
    message: String,
    frame_idx: usize,
}

impl SimpleSpinner {
    pub fn new(message: &str) -> Self {
        SimpleSpinner {
            message: message.to_string(),
            frame_idx: 0,
        }
    }
    
    pub fn tick(&mut self) -> io::Result<()> {
        let frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
        
        print!("\r{} {}  ", 
               frames[self.frame_idx].bright_cyan(), 
               self.message.dimmed());
        io::stdout().flush()?;
        
        self.frame_idx = (self.frame_idx + 1) % frames.len();
        Ok(())
    }
    
    pub fn clear(&self) -> io::Result<()> {
        print!("\r{}", " ".repeat(self.message.len() + 10));
        print!("\r");
        io::stdout().flush()?;
        Ok(())
    }
}