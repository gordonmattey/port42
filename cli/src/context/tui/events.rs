// Event handling for TUI

use anyhow::Result;
use crossterm::event::{self, Event as CrosstermEvent, KeyEvent};
use std::time::Duration;
use tokio::sync::mpsc;
use tokio::time::interval;

#[derive(Debug, Clone)]
pub enum Event {
    /// Terminal tick for refreshing data
    Tick,
    /// Key press event
    Key(KeyEvent),
    /// Terminal resize event
    Resize(u16, u16),
    /// Quit signal
    Quit,
}

pub struct EventHandler {
    rx: mpsc::UnboundedReceiver<Event>,
    _tx: mpsc::UnboundedSender<Event>,
}

impl EventHandler {
    pub fn new(tick_rate: Duration) -> Self {
        let (tx, rx) = mpsc::unbounded_channel();
        let tx_clone = tx.clone();

        // Spawn task to handle crossterm events
        tokio::spawn(async move {
            loop {
                if event::poll(Duration::from_millis(50)).unwrap_or(false) {
                    if let Ok(evt) = event::read() {
                        match evt {
                            CrosstermEvent::Key(key) => {
                                // Check for quit
                                if key.code == event::KeyCode::Char('q')
                                    && key.modifiers == event::KeyModifiers::NONE
                                {
                                    let _ = tx_clone.send(Event::Quit);
                                    break;
                                }
                                let _ = tx_clone.send(Event::Key(key));
                            }
                            CrosstermEvent::Resize(width, height) => {
                                let _ = tx_clone.send(Event::Resize(width, height));
                            }
                            _ => {}
                        }
                    }
                }
            }
        });

        // Spawn task for tick events
        let tx_clone = tx.clone();
        tokio::spawn(async move {
            let mut ticker = interval(tick_rate);
            loop {
                ticker.tick().await;
                if tx_clone.send(Event::Tick).is_err() {
                    break;
                }
            }
        });

        Self { rx, _tx: tx }
    }

    pub async fn next(&mut self) -> Result<Event> {
        self.rx
            .recv()
            .await
            .ok_or_else(|| anyhow::anyhow!("Event channel closed"))
    }
}