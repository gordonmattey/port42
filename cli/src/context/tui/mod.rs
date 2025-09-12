// TUI module for context watch mode

pub mod app;
pub mod events;
pub mod ui;

pub use app::App;
pub use events::{Event, EventHandler};

use anyhow::Result;
use crossterm::{
    event::{DisableMouseCapture, EnableMouseCapture},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::CrosstermBackend,
    Terminal,
};
use std::io;

/// Initialize the terminal for TUI mode
pub fn init_terminal() -> Result<Terminal<CrosstermBackend<io::Stdout>>> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let terminal = Terminal::new(backend)?;
    Ok(terminal)
}

/// Restore the terminal to normal mode
pub fn restore_terminal(terminal: &mut Terminal<CrosstermBackend<io::Stdout>>) -> Result<()> {
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;
    Ok(())
}

/// Run the TUI application
pub async fn run_app(
    terminal: &mut Terminal<CrosstermBackend<io::Stdout>>,
    app: &mut App,
    event_handler: &mut EventHandler,
) -> Result<()> {
    // Set up panic handler to restore terminal
    let original_hook = std::panic::take_hook();
    std::panic::set_hook(Box::new(move |panic| {
        let _ = disable_raw_mode();
        let _ = execute!(
            io::stdout(),
            LeaveAlternateScreen,
            DisableMouseCapture
        );
        original_hook(panic);
    }));
    
    let result = run_app_loop(terminal, app, event_handler).await;
    
    // Restore original panic handler
    let _ = std::panic::take_hook();
    
    result
}

async fn run_app_loop(
    terminal: &mut Terminal<CrosstermBackend<io::Stdout>>,
    app: &mut App,
    event_handler: &mut EventHandler,
) -> Result<()> {
    loop {
        // Draw the UI
        terminal.draw(|f| ui::draw(f, app))?;

        // Handle events
        match event_handler.next().await? {
            Event::Quit => {
                return Ok(());
            }
            event => {
                // Check if app wants to quit
                if app.handle_event(event)? {
                    return Ok(());
                }
            }
        }
    }
}