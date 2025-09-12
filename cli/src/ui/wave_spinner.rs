//! Wave emoji spinner for swimming mode
//! 
//! Shows an animated wave while waiting for responses

use std::io::{self, Write};
use std::sync::mpsc::{self, Sender};
use std::thread;
use std::time::Duration;
use crossterm::{cursor, execute};

pub struct WaveSpinner {
    handle: Option<thread::JoinHandle<()>>,
    stop_sender: Option<Sender<()>>,
}

impl WaveSpinner {
    pub fn new() -> Self {
        let (tx, rx) = mpsc::channel();
        
        let handle = thread::spawn(move || {
            // Alternate between wave and space for flashing effect
            let frames = ["🌊", "  "];
            let mut frame_idx = 0;
            
            // Hide cursor
            let _ = execute!(io::stdout(), cursor::Hide);
            
            loop {
                // Check if we should stop
                if rx.try_recv().is_ok() {
                    break;
                }
                
                // Print wave frame
                print!("\r{}  ", frames[frame_idx]);
                let _ = io::stdout().flush();
                
                frame_idx = (frame_idx + 1) % frames.len();
                
                // Sleep for animation speed (slower for wave effect)
                thread::sleep(Duration::from_millis(500));
            }
            
            // Clear the line and show cursor again
            print!("\r    \r");
            let _ = execute!(io::stdout(), cursor::Show);
            let _ = io::stdout().flush();
        });
        
        Self {
            handle: Some(handle),
            stop_sender: Some(tx),
        }
    }
    
    pub fn stop(&mut self) {
        // Send stop signal
        if let Some(sender) = self.stop_sender.take() {
            let _ = sender.send(());
        }
        
        // Wait for thread to finish
        if let Some(handle) = self.handle.take() {
            let _ = handle.join();
        }
    }
}

impl Drop for WaveSpinner {
    fn drop(&mut self) {
        // Ensure cleanup on drop
        if let Some(sender) = self.stop_sender.take() {
            let _ = sender.send(());
        }
        if let Some(handle) = self.handle.take() {
            let _ = handle.join();
        }
    }
}