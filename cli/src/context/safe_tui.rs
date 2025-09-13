// Safe TUI implementation with guaranteed terminal restoration

use anyhow::Result;
use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode, KeyModifiers},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::CrosstermBackend,
    layout::{Alignment, Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, List, ListItem, Paragraph},
    Frame, Terminal,
};
use std::{
    io::{self, Stdout},
    panic::{self, PanicInfo},
    sync::{Arc, Mutex},
    time::{Duration, Instant},
};

use crate::client::DaemonClient;
use crate::context::ContextData;

/// Guard that ensures terminal is always restored
struct TerminalGuard {
    // Use Arc<Mutex> to allow access in panic handler
    restored: Arc<Mutex<bool>>,
}

impl TerminalGuard {
    fn new() -> Result<Self> {
        // Try to enable raw mode and alternate screen - let Ratatui handle terminal detection
        enable_raw_mode().map_err(|e| anyhow::anyhow!("Failed to enable raw mode: {}", e))?;
        execute!(io::stdout(), EnterAlternateScreen, EnableMouseCapture)
            .map_err(|e| anyhow::anyhow!("Failed to setup terminal: {}", e))?;
        
        let restored = Arc::new(Mutex::new(false));
        let restored_clone = restored.clone();
        
        // Install panic hook to restore terminal
        let original_hook = panic::take_hook();
        panic::set_hook(Box::new(move |info: &PanicInfo| {
            // Restore terminal before panic
            Self::force_restore(&restored_clone);
            // Call original panic handler
            original_hook(info);
        }));
        
        Ok(Self { restored })
    }
    
    fn force_restore(restored: &Arc<Mutex<bool>>) {
        if let Ok(mut restored_flag) = restored.lock() {
            if !*restored_flag {
                // Best effort restoration
                let _ = disable_raw_mode();
                let _ = execute!(
                    io::stdout(),
                    LeaveAlternateScreen,
                    DisableMouseCapture
                );
                *restored_flag = true;
            }
        }
    }
    
    fn restore(&self) {
        Self::force_restore(&self.restored);
    }
}

impl Drop for TerminalGuard {
    fn drop(&mut self) {
        self.restore();
    }
}

/// Safe terminal wrapper that guarantees cleanup
pub struct SafeTerminal {
    terminal: Terminal<CrosstermBackend<Stdout>>,
    _guard: TerminalGuard,  // Dropped last, restoring terminal
}

impl SafeTerminal {
    pub fn new() -> Result<Self> {
        let guard = TerminalGuard::new()?;
        let backend = CrosstermBackend::new(io::stdout());
        let terminal = Terminal::new(backend)?;
        
        Ok(Self {
            terminal,
            _guard: guard,
        })
    }
    
    pub fn draw<F>(&mut self, f: F) -> Result<()>
    where
        F: FnOnce(&mut Frame),
    {
        self.terminal.draw(f)?;
        Ok(())
    }
}

/// Rate limiter to prevent terminal overwhelm
struct RateLimiter {
    last_update: Instant,
    min_interval: Duration,
}

impl RateLimiter {
    fn new(min_interval_ms: u64) -> Self {
        Self {
            last_update: Instant::now(),
            min_interval: Duration::from_millis(min_interval_ms),
        }
    }
    
    fn should_update(&mut self) -> bool {
        if self.last_update.elapsed() >= self.min_interval {
            self.last_update = Instant::now();
            true
        } else {
            false
        }
    }
}

/// Activity record for display
#[derive(Debug, Clone)]
struct Activity {
    timestamp: chrono::DateTime<chrono::Local>,
    activity_type: String,
    description: String,
    color: Color,
}

/// Main application state
pub struct App {
    activities: Vec<Activity>,
    selected: usize,
    scroll_offset: usize,
    viewport_height: usize,
    should_quit: bool,
    daemon_client: DaemonClient,
    last_error: Option<String>,
    rate_limiter: RateLimiter,
    active_session: Option<String>,
    active_agent: Option<String>,
}

impl App {
    pub fn new(daemon_client: DaemonClient) -> Self {
        Self {
            activities: Vec::new(),
            selected: 0,
            scroll_offset: 0,
            viewport_height: 20,
            should_quit: false,
            daemon_client,
            last_error: None,
            rate_limiter: RateLimiter::new(500), // Not used anymore, kept for compatibility
            active_session: None,
            active_agent: None,
        }
    }
    
    fn handle_key(&mut self, code: KeyCode, modifiers: KeyModifiers) -> Result<()> {
        // Ctrl+C always quits
        if code == KeyCode::Char('c') && modifiers == KeyModifiers::CONTROL {
            self.should_quit = true;
            return Ok(());
        }
        
        match code {
            KeyCode::Char('q') => self.should_quit = true,
            KeyCode::Up | KeyCode::Char('k') => self.move_up(),
            KeyCode::Down | KeyCode::Char('j') => self.move_down(),
            KeyCode::PageUp => self.page_up(),
            KeyCode::PageDown => self.page_down(),
            KeyCode::Home => self.go_to_top(),
            KeyCode::End => self.go_to_bottom(),
            _ => {}
        }
        
        Ok(())
    }
    
    fn move_up(&mut self) {
        if self.selected > 0 {
            self.selected -= 1;
            if self.selected < self.scroll_offset {
                self.scroll_offset = self.selected;
            }
        }
    }
    
    fn move_down(&mut self) {
        let max_index = self.activities.len().saturating_sub(1);
        if self.selected < max_index {
            self.selected += 1;
            if self.selected >= self.scroll_offset + self.viewport_height {
                self.scroll_offset = self.selected - self.viewport_height + 1;
            }
        }
    }
    
    fn page_up(&mut self) {
        let page_size = self.viewport_height.saturating_sub(1);
        self.selected = self.selected.saturating_sub(page_size);
        self.scroll_offset = self.scroll_offset.saturating_sub(page_size);
    }
    
    fn page_down(&mut self) {
        let max_index = self.activities.len().saturating_sub(1);
        let page_size = self.viewport_height.saturating_sub(1);
        self.selected = (self.selected + page_size).min(max_index);
        
        if self.selected >= self.scroll_offset + self.viewport_height {
            self.scroll_offset = self.selected - self.viewport_height + 1;
        }
    }
    
    fn go_to_top(&mut self) {
        self.selected = 0;
        self.scroll_offset = 0;
    }
    
    fn go_to_bottom(&mut self) {
        let max_index = self.activities.len().saturating_sub(1);
        self.selected = max_index;
        self.scroll_offset = max_index.saturating_sub(self.viewport_height - 1);
    }
    
    fn refresh_data(&mut self) -> Result<()> {
        // Remove rate limiter check - the main loop already controls refresh timing
        // The rate limiter was causing conflicts with the main refresh interval
        
        // Try to get context from daemon
        use crate::protocol::DaemonRequest;
        
        let request = DaemonRequest {
            request_type: "context".to_string(),
            id: format!("watch-{}", chrono::Utc::now().timestamp_millis()),
            payload: serde_json::json!({}),
            references: None,
            session_context: None,
            user_prompt: None,
        };
        
        match self.daemon_client.request(request) {
            Ok(response) => {
                if let Some(data) = response.data {
                    if let Ok(context) = serde_json::from_value::<ContextData>(data) {
                        self.process_context(context);
                        self.last_error = None;
                    } else {
                        self.last_error = Some("Failed to parse context data".to_string());
                    }
                } else {
                    self.last_error = Some("No data in daemon response".to_string());
                }
            }
            Err(e) => {
                self.last_error = Some(format!("Daemon error: {}", e));
            }
        }
        
        Ok(())
    }
    
    fn process_context(&mut self, context: ContextData) {
        // Update active session info
        if let Some(session) = context.active_session.as_ref() {
            self.active_session = Some(session.id.clone());
            self.active_agent = Some(session.agent.clone());
        } else {
            self.active_session = None;
            self.active_agent = None;
        }
        
        // Clear and rebuild activities
        self.activities.clear();
        
        // Add active session as an activity if present
        if let Some(ref session) = context.active_session {
            self.activities.push(Activity {
                timestamp: session.last_activity.with_timezone(&chrono::Local),
                activity_type: "SESSION".to_string(),
                description: format!("Active: {} ({} msgs)", session.agent, session.message_count),
                color: Color::Cyan,
            });
        }
        
        // Add recent commands
        for cmd in context.recent_commands {
            self.activities.push(Activity {
                timestamp: cmd.timestamp.with_timezone(&chrono::Local),
                activity_type: "COMMAND".to_string(),
                description: cmd.command,
                color: Color::Blue,
            });
        }
        
        // Add created tools
        for tool in context.created_tools {
            self.activities.push(Activity {
                timestamp: tool.created_at.with_timezone(&chrono::Local),
                activity_type: "TOOL".to_string(),
                description: format!("Created: {}", tool.name),
                color: Color::Magenta,
            });
        }
        
        // Add memory accesses
        for mem in context.accessed_memories {
            // Memory accesses don't have timestamps in the data, use current time
            // This is a limitation we should fix in the daemon later
            self.activities.push(Activity {
                timestamp: chrono::Local::now(),
                activity_type: "MEMORY".to_string(),
                description: format!("Accessed: {}", mem.path),
                color: Color::Green,
            });
        }
        
        // Sort by timestamp (newest first)
        self.activities.sort_by(|a, b| b.timestamp.cmp(&a.timestamp));
    }
    
    fn render(&self, frame: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3),  // Header
                Constraint::Min(0),     // Body
                Constraint::Length(3),  // Footer
            ])
            .split(frame.size());
        
        self.render_header(frame, chunks[0]);
        self.render_activities(frame, chunks[1]);
        self.render_footer(frame, chunks[2]);
    }
    
    fn render_header(&self, frame: &mut Frame, area: Rect) {
        let header_text = if let Some(err) = &self.last_error {
            vec![
                Span::styled("⚠️ ", Style::default().fg(Color::Red)),
                Span::styled(err, Style::default().fg(Color::Red)),
            ]
        } else {
            let mut spans = vec![
                Span::styled("🔍 ", Style::default()),
                Span::styled(
                    "Port42 Context Monitor",
                    Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD),
                ),
                Span::raw(" │ "),
                Span::styled(
                    format!("{} activities", self.activities.len()),
                    Style::default().fg(Color::Yellow),
                ),
            ];
            
            // Show active session if present
            if let Some(ref session_id) = self.active_session {
                spans.push(Span::raw(" │ "));
                
                // Show agent if present
                if let Some(ref agent) = self.active_agent {
                    spans.push(Span::styled(
                        agent.clone(),
                        Style::default().fg(Color::Cyan),
                    ));
                    spans.push(Span::raw(" "));
                }
                
                // Show full session ID
                spans.push(Span::styled(
                    session_id.clone(),
                    Style::default().fg(Color::Blue),
                ));
            }
            
            spans
        };
        
        let header = Paragraph::new(Line::from(header_text))
            .block(
                Block::default()
                    .borders(Borders::BOTTOM)
                    .border_style(Style::default().fg(Color::DarkGray)),
            )
            .alignment(Alignment::Center);
        
        frame.render_widget(header, area);
    }
    
    fn render_activities(&self, frame: &mut Frame, area: Rect) {
        // Update viewport height
        let viewport_height = area.height as usize;
        
        // If no activities, show a helpful message
        if self.activities.is_empty() {
            let message = Paragraph::new(
                Line::from(vec![
                    Span::styled(
                        "No recent activity. Run some Port42 commands to see them here!",
                        Style::default().fg(Color::DarkGray).add_modifier(Modifier::ITALIC),
                    ),
                ])
            )
            .block(Block::default().borders(Borders::NONE))
            .alignment(Alignment::Center);
            
            frame.render_widget(message, area);
            return;
        }
        
        let items: Vec<ListItem> = self.activities
            .iter()
            .skip(self.scroll_offset)
            .take(viewport_height)
            .enumerate()
            .map(|(i, activity)| {
                let is_selected = i + self.scroll_offset == self.selected;
                
                let timestamp_style = if is_selected {
                    Style::default().fg(Color::White)
                } else {
                    Style::default().fg(Color::Gray)
                };
                
                let spans = vec![
                    Span::styled(
                        format!("{:<8} ", activity.timestamp.format("%H:%M:%S").to_string()),
                        timestamp_style,
                    ),
                    Span::styled(
                        format!("{:<8} ", activity.activity_type),
                        Style::default().fg(activity.color),
                    ),
                    Span::raw(&activity.description),
                ];
                
                let style = if is_selected {
                    Style::default().bg(Color::DarkGray).add_modifier(Modifier::BOLD)
                } else {
                    Style::default()
                };
                
                ListItem::new(Line::from(spans)).style(style)
            })
            .collect();
        
        let list = List::new(items).block(Block::default().borders(Borders::NONE));
        frame.render_widget(list, area);
    }
    
    fn render_footer(&self, frame: &mut Frame, area: Rect) {
        let keybinds = vec![
            ("q/Ctrl+C", "quit"),
            ("↑↓", "navigate"),
            ("PgUp/PgDn", "page"),
            ("Home/End", "top/bottom"),
        ];
        
        let keybind_text: Vec<Span> = keybinds
            .iter()
            .flat_map(|(key, desc)| {
                vec![
                    Span::styled(
                        format!("[{}]", key),
                        Style::default().fg(Color::Yellow).add_modifier(Modifier::BOLD),
                    ),
                    Span::styled(format!("{} ", desc), Style::default().fg(Color::White)),
                ]
            })
            .collect();
        
        let footer = Paragraph::new(Line::from(keybind_text))
            .block(
                Block::default()
                    .borders(Borders::TOP)
                    .border_style(Style::default().fg(Color::DarkGray)),
            )
            .alignment(Alignment::Center);
        
        frame.render_widget(footer, area);
    }
}

/// Main entry point for safe TUI
pub fn run_safe_watch(mut daemon_client: DaemonClient, refresh_ms: u64) -> Result<()> {
    // Create safe terminal (will auto-restore on drop)
    let mut terminal = SafeTerminal::new()?;
    
    // Create app
    let mut app = App::new(daemon_client);
    
    // Timing for refresh
    let refresh_interval = Duration::from_millis(refresh_ms);
    let mut last_refresh = Instant::now();
    
    // Initial data fetch
    app.refresh_data()?;
    
    // Main synchronous event loop
    loop {
        // Check if it's time to refresh data BEFORE rendering
        if last_refresh.elapsed() >= refresh_interval {
            app.refresh_data()?;
            last_refresh = Instant::now();
        }
        
        // Render UI with current data
        terminal.draw(|f| app.render(f))?;
        
        // Check if we should quit
        if app.should_quit {
            break;
        }
        
        // Poll for events with short timeout for responsiveness
        if event::poll(Duration::from_millis(50))? {
            match event::read()? {
                Event::Key(key) => {
                    app.handle_key(key.code, key.modifiers)?;
                }
                Event::Resize(_, height) => {
                    app.viewport_height = height.saturating_sub(6) as usize;
                }
                _ => {}
            }
        }
    }
    
    Ok(())
    // Terminal automatically restored when SafeTerminal drops
}