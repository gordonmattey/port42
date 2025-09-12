// New TUI-based watch mode implementation

use anyhow::Result;
use std::time::Duration;
use crate::client::DaemonClient;

use super::tui::{self, App, EventHandler};

pub struct WatchMode {
    client: DaemonClient,
    refresh_rate: Duration,
}

impl WatchMode {
    pub fn new(client: DaemonClient, refresh_ms: u64) -> Self {
        Self {
            client,
            refresh_rate: Duration::from_millis(refresh_ms),
        }
    }

    pub async fn run(self) -> Result<()> {
        // Initialize terminal
        let mut terminal = tui::init_terminal()?;
        
        // Create app state
        let mut app = App::new(self.client);
        
        // Create event handler
        let mut event_handler = EventHandler::new(self.refresh_rate);
        
        // Run the app
        let result = tui::run_app(&mut terminal, &mut app, &mut event_handler).await;
        
        // Restore terminal
        tui::restore_terminal(&mut terminal)?;
        
        result
    }
}