// TUI Application State

use anyhow::Result;
use std::time::Instant;
use crate::client::DaemonClient;
use super::Event;
use chrono::{DateTime, Utc, Local};

#[derive(Debug, Clone)]
pub struct ActivityRecord {
    pub timestamp: String,
    pub activity_type: ActivityType,
    pub description: String,
    pub details: Option<String>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum ActivityType {
    Command,
    Memory,
    FileAccess,
    ToolUsage,
    Error,
}

impl ActivityType {
    pub fn as_str(&self) -> &str {
        match self {
            ActivityType::Command => "COMMAND",
            ActivityType::Memory => "MEMORY",
            ActivityType::FileAccess => "ACCESS",
            ActivityType::ToolUsage => "TOOL",
            ActivityType::Error => "ERROR",
        }
    }

    pub fn color(&self) -> ratatui::style::Color {
        use ratatui::style::Color;
        match self {
            ActivityType::Command => Color::Blue,
            ActivityType::Memory => Color::Green,
            ActivityType::FileAccess => Color::Cyan,
            ActivityType::ToolUsage => Color::Magenta,
            ActivityType::Error => Color::LightRed,
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub enum FilterMode {
    All,
    Commands,
    Memory,
    FileAccess,
    ToolUsage,
    Search(String),
}

pub struct App {
    // Activity management
    pub activities: Vec<ActivityRecord>,
    pub filtered_activities: Vec<ActivityRecord>,
    max_activities: usize,
    
    // UI state
    pub selected_index: usize,
    pub scroll_offset: usize,
    pub viewport_height: usize,
    
    // Filtering
    pub filter_mode: FilterMode,
    pub filter_text: String,
    pub is_filtering: bool,
    
    // View options
    pub show_details: bool,
    pub auto_scroll: bool,
    pub show_timestamps: bool,
    pub show_help: bool,
    
    // Stats
    pub total_commands: usize,
    pub commands_per_minute: f64,
    pub active_session: Option<String>,
    pub last_refresh: Instant,
    
    // Connection
    pub daemon_client: DaemonClient,
}

impl App {
    pub fn new(daemon_client: DaemonClient) -> Self {
        let mut app = Self {
            activities: Vec::new(),
            filtered_activities: Vec::new(),
            max_activities: 1000,
            selected_index: 0,
            scroll_offset: 0,
            viewport_height: 20,
            filter_mode: FilterMode::All,
            filter_text: String::new(),
            is_filtering: false,
            show_details: false,
            auto_scroll: true,
            show_timestamps: true,
            show_help: false,
            total_commands: 0,
            commands_per_minute: 0.0,
            active_session: None,
            last_refresh: Instant::now(),
            daemon_client,
        };
        
        // Add some demo data for testing
        app.add_demo_activities();
        app
    }
    
    fn add_demo_activities(&mut self) {
        // Add some demo activities for testing without daemon
        let demo_activities = vec![
            ActivityRecord {
                timestamp: "12:34:56".to_string(),
                activity_type: ActivityType::Command,
                description: "port42 status".to_string(),
                details: Some("Checking daemon status".to_string()),
            },
            ActivityRecord {
                timestamp: "12:34:50".to_string(),
                activity_type: ActivityType::Memory,
                description: "Created session cli-test".to_string(),
                details: Some("New AI conversation started".to_string()),
            },
            ActivityRecord {
                timestamp: "12:34:45".to_string(),
                activity_type: ActivityType::FileAccess,
                description: "Read /tools/".to_string(),
                details: Some("Browsed 15 tools".to_string()),
            },
            ActivityRecord {
                timestamp: "12:34:40".to_string(),
                activity_type: ActivityType::ToolUsage,
                description: "git-haiku".to_string(),
                details: Some("Generated commit haiku".to_string()),
            },
            ActivityRecord {
                timestamp: "12:34:35".to_string(),
                activity_type: ActivityType::Command,
                description: "port42 possess @ai-engineer".to_string(),
                details: Some("Started AI session".to_string()),
            },
        ];
        
        for activity in demo_activities {
            self.activities.push(activity);
        }
        self.update_filter();
    }

    pub fn handle_event(&mut self, event: Event) -> Result<bool> {
        match event {
            Event::Tick => {
                self.refresh_activities()?;
            }
            Event::Key(key) => {
                return self.handle_key_event(key);
            }
            Event::Resize(_width, height) => {
                self.viewport_height = height.saturating_sub(7) as usize;
            }
            Event::Quit => {
                return Ok(true);  // Signal to quit
            }
            _ => {}
        }
        Ok(false)  // Don't quit
    }

    fn handle_key_event(&mut self, key: crossterm::event::KeyEvent) -> Result<bool> {
        use crossterm::event::{KeyCode, KeyModifiers};

        // Handle Ctrl+C to quit from anywhere
        if key.code == KeyCode::Char('c') && key.modifiers == KeyModifiers::CONTROL {
            return Ok(true);  // Quit
        }

        match key.code {
            KeyCode::Char('q') if !self.is_filtering => {
                return Ok(true);  // Quit
            }
            KeyCode::Up | KeyCode::Char('k') if !self.is_filtering => {
                self.move_selection_up();
            }
            KeyCode::Down | KeyCode::Char('j') if !self.is_filtering => {
                self.move_selection_down();
            }
            KeyCode::PageUp | KeyCode::Char('u') if !self.is_filtering => {
                self.page_up();
            }
            KeyCode::PageDown | KeyCode::Char('d') if !self.is_filtering => {
                self.page_down();
            }
            KeyCode::Home | KeyCode::Char('g') if !self.is_filtering => {
                self.go_to_top();
            }
            KeyCode::End | KeyCode::Char('G') if !self.is_filtering => {
                self.go_to_bottom();
            }
            KeyCode::Char(' ') if !self.is_filtering => {
                self.show_details = !self.show_details;
            }
            KeyCode::Char('a') if !self.is_filtering => {
                self.auto_scroll = !self.auto_scroll;
            }
            KeyCode::Char('t') if !self.is_filtering => {
                self.show_timestamps = !self.show_timestamps;
            }
            KeyCode::Char('f') if !self.is_filtering => {
                self.cycle_filter_mode();
            }
            KeyCode::Char('/') if !self.is_filtering => {
                self.start_search();
            }
            KeyCode::Char('c') if !self.is_filtering => {
                self.clear_activities();
            }
            KeyCode::Char('?') if !self.is_filtering => {
                self.show_help = !self.show_help;
            }
            KeyCode::Esc if self.is_filtering => {
                self.cancel_search();
            }
            KeyCode::Enter if self.is_filtering => {
                self.apply_search();
            }
            KeyCode::Backspace if self.is_filtering => {
                self.filter_text.pop();
                self.update_filter();
            }
            KeyCode::Char(c) if self.is_filtering => {
                self.filter_text.push(c);
                self.update_filter();
            }
            _ => {}
        }
        Ok(false)  // Don't quit
    }

    fn refresh_activities(&mut self) -> Result<()> {
        // Try to fetch context from daemon
        match self.fetch_daemon_context() {
            Ok(new_activities) => {
                // Add new activities that we haven't seen
                for activity in new_activities {
                    if !self.activity_exists(&activity) {
                        self.add_activity(activity);
                    }
                }
            }
            Err(e) => {
                // If daemon is down, add an error activity (but only once)
                if !self.activities.iter().any(|a| {
                    a.activity_type == ActivityType::Error 
                    && a.description.contains("Daemon connection failed")
                }) {
                    self.add_activity(ActivityRecord {
                        timestamp: chrono::Local::now().format("%H:%M:%S").to_string(),
                        activity_type: ActivityType::Error,
                        description: "Daemon connection failed".to_string(),
                        details: Some(format!("Error: {}", e)),
                    });
                }
            }
        }
        
        Ok(())
    }
    
    fn fetch_daemon_context(&mut self) -> Result<Vec<ActivityRecord>> {
        use crate::protocol::DaemonRequest;
        
        // Create context request
        let request = DaemonRequest {
            request_type: "context".to_string(),
            id: format!("context-{}", std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_millis()),
            payload: serde_json::json!({}),
            references: None,
            session_context: None,
            user_prompt: None,
        };
        
        // Try to get response from daemon
        let response = self.daemon_client.request(request)?;
        
        // Parse the context data
        let context_data: crate::context::ContextData = serde_json::from_value(
            response.data.ok_or_else(|| anyhow::anyhow!("No data in response"))?
        )?;
        
        // Convert context data to activities
        let mut activities = Vec::new();
        
        // Add recent commands as activities
        for cmd in &context_data.recent_commands {
            activities.push(ActivityRecord {
                timestamp: cmd.timestamp.format("%H:%M:%S").to_string(),
                activity_type: ActivityType::Command,
                description: cmd.command.clone(),
                details: Some(format!("Exit code: {}", cmd.exit_code)),
            });
        }
        
        // Add created tools as activities
        for tool in &context_data.created_tools {
            activities.push(ActivityRecord {
                timestamp: tool.created_at.format("%H:%M:%S").to_string(),
                activity_type: ActivityType::ToolUsage,
                description: format!("Created tool: {}", tool.name),
                details: Some(format!("Type: {}", tool.tool_type)),
            });
        }
        
        // Add memory accesses as activities
        for mem in &context_data.accessed_memories {
            activities.push(ActivityRecord {
                timestamp: chrono::Local::now().format("%H:%M:%S").to_string(),
                activity_type: ActivityType::FileAccess,
                description: format!("Accessed {}", mem.path),
                details: mem.display_name.clone(),
            });
        }
        
        // Add active session info if present
        if let Some(session) = &context_data.active_session {
            self.active_session = Some(session.id.clone());
            
            // Update stats
            self.total_commands = context_data.recent_commands.len();
            let now = Utc::now();
            let elapsed = (now - session.start_time).num_seconds() as f64 / 60.0;
            if elapsed > 0.0 {
                self.commands_per_minute = self.total_commands as f64 / elapsed;
            }
        }
        
        Ok(activities)
    }
    
    fn activity_exists(&self, activity: &ActivityRecord) -> bool {
        self.activities.iter().any(|a| {
            a.timestamp == activity.timestamp 
            && a.description == activity.description
        })
    }

    pub fn add_activity(&mut self, activity: ActivityRecord) {
        // Add to ring buffer
        if self.activities.len() >= self.max_activities {
            self.activities.remove(0);
        }
        
        self.activities.push(activity);
        
        // Update filtered view
        self.update_filter();
        
        // Auto-scroll to bottom if enabled
        if self.auto_scroll {
            self.go_to_bottom();
        }
        
        // Update stats
        self.update_stats();
    }

    fn update_filter(&mut self) {
        self.filtered_activities = match &self.filter_mode {
            FilterMode::All => self.activities.clone(),
            FilterMode::Commands => {
                self.activities
                    .iter()
                    .filter(|a| a.activity_type == ActivityType::Command)
                    .cloned()
                    .collect()
            }
            FilterMode::Memory => {
                self.activities
                    .iter()
                    .filter(|a| a.activity_type == ActivityType::Memory)
                    .cloned()
                    .collect()
            }
            FilterMode::FileAccess => {
                self.activities
                    .iter()
                    .filter(|a| a.activity_type == ActivityType::FileAccess)
                    .cloned()
                    .collect()
            }
            FilterMode::ToolUsage => {
                self.activities
                    .iter()
                    .filter(|a| a.activity_type == ActivityType::ToolUsage)
                    .cloned()
                    .collect()
            }
            FilterMode::Search(query) => {
                self.activities
                    .iter()
                    .filter(|a| {
                        a.description.to_lowercase().contains(&query.to_lowercase())
                            || a.details
                                .as_ref()
                                .map(|d| d.to_lowercase().contains(&query.to_lowercase()))
                                .unwrap_or(false)
                    })
                    .cloned()
                    .collect()
            }
        };
    }

    fn move_selection_up(&mut self) {
        if self.selected_index > 0 {
            self.selected_index -= 1;
            if self.selected_index < self.scroll_offset {
                self.scroll_offset = self.selected_index;
            }
        }
    }

    fn move_selection_down(&mut self) {
        let max_index = self.filtered_activities.len().saturating_sub(1);
        if self.selected_index < max_index {
            self.selected_index += 1;
            if self.selected_index >= self.scroll_offset + self.viewport_height {
                self.scroll_offset = self.selected_index - self.viewport_height + 1;
            }
        }
    }

    fn page_up(&mut self) {
        let page_size = self.viewport_height.saturating_sub(1);
        self.selected_index = self.selected_index.saturating_sub(page_size);
        self.scroll_offset = self.scroll_offset.saturating_sub(page_size);
    }

    fn page_down(&mut self) {
        let max_index = self.filtered_activities.len().saturating_sub(1);
        let page_size = self.viewport_height.saturating_sub(1);
        self.selected_index = (self.selected_index + page_size).min(max_index);
        
        if self.selected_index >= self.scroll_offset + self.viewport_height {
            self.scroll_offset = self.selected_index - self.viewport_height + 1;
        }
    }

    fn go_to_top(&mut self) {
        self.selected_index = 0;
        self.scroll_offset = 0;
    }

    fn go_to_bottom(&mut self) {
        let max_index = self.filtered_activities.len().saturating_sub(1);
        self.selected_index = max_index;
        self.scroll_offset = max_index.saturating_sub(self.viewport_height - 1);
    }

    fn cycle_filter_mode(&mut self) {
        self.filter_mode = match self.filter_mode {
            FilterMode::All => FilterMode::Commands,
            FilterMode::Commands => FilterMode::Memory,
            FilterMode::Memory => FilterMode::FileAccess,
            FilterMode::FileAccess => FilterMode::ToolUsage,
            FilterMode::ToolUsage => FilterMode::All,
            FilterMode::Search(_) => FilterMode::All,
        };
        self.update_filter();
    }

    fn start_search(&mut self) {
        self.is_filtering = true;
        self.filter_text.clear();
    }

    fn cancel_search(&mut self) {
        self.is_filtering = false;
        self.filter_text.clear();
        self.filter_mode = FilterMode::All;
        self.update_filter();
    }

    fn apply_search(&mut self) {
        self.is_filtering = false;
        if !self.filter_text.is_empty() {
            self.filter_mode = FilterMode::Search(self.filter_text.clone());
            self.update_filter();
        }
    }

    fn clear_activities(&mut self) {
        self.activities.clear();
        self.filtered_activities.clear();
        self.selected_index = 0;
        self.scroll_offset = 0;
        self.total_commands = 0;
    }

    fn update_stats(&mut self) {
        // Count commands
        self.total_commands = self.activities
            .iter()
            .filter(|a| a.activity_type == ActivityType::Command)
            .count();
        
        // Calculate rate
        let elapsed = self.last_refresh.elapsed().as_secs_f64() / 60.0;
        if elapsed > 0.0 {
            self.commands_per_minute = self.total_commands as f64 / elapsed;
        }
    }
}