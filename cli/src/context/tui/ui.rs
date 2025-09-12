// UI rendering for TUI

use ratatui::{
    layout::{Alignment, Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, List, ListItem, Paragraph},
    Frame,
};

use super::app::{App, FilterMode};

pub fn draw(f: &mut Frame, app: &App) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),  // Header
            Constraint::Min(0),     // Body
            Constraint::Length(3),  // Footer
        ])
        .split(f.size());

    draw_header(f, chunks[0], app);
    
    if app.show_help {
        draw_help(f, chunks[1]);
    } else {
        draw_activities(f, chunks[1], app);
    }
    
    draw_footer(f, chunks[2], app);
}

fn draw_header(f: &mut Frame, area: Rect, app: &App) {
    let header_text = vec![
        Span::styled("🔍 ", Style::default()),
        Span::styled(
            "Port42 Context Monitor",
            Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD),
        ),
        Span::raw(" │ "),
        Span::styled(
            format!("{} activities", app.filtered_activities.len()),
            Style::default().fg(Color::Yellow),
        ),
        Span::raw(" │ "),
        Span::styled(
            format!("{:.1} cmd/m", app.commands_per_minute),
            Style::default().fg(Color::Green),
        ),
        Span::raw(" │ "),
        Span::styled(
            format_filter_mode(&app.filter_mode),
            Style::default().fg(Color::Magenta),
        ),
    ];

    let header = Paragraph::new(Line::from(header_text))
        .block(
            Block::default()
                .borders(Borders::BOTTOM)
                .border_style(Style::default().fg(Color::DarkGray)),
        )
        .alignment(Alignment::Center);

    f.render_widget(header, area);
}

fn draw_activities(f: &mut Frame, area: Rect, app: &App) {
    let items: Vec<ListItem> = app
        .filtered_activities
        .iter()
        .enumerate()
        .skip(app.scroll_offset)
        .take(app.viewport_height)
        .map(|(i, activity)| {
            let mut spans = vec![];
            
            // Add timestamp if enabled
            if app.show_timestamps {
                spans.push(Span::styled(
                    format!("{:<8} ", activity.timestamp),
                    Style::default().fg(Color::DarkGray),
                ));
            }
            
            // Add activity type with color
            spans.push(Span::styled(
                format!("{:<8} ", activity.activity_type.as_str()),
                Style::default().fg(activity.activity_type.color()),
            ));
            
            // Add description
            spans.push(Span::raw(format!("{:<30} ", activity.description)));
            
            // Add details if available
            if let Some(details) = &activity.details {
                spans.push(Span::styled(
                    details,
                    Style::default().fg(Color::Gray),
                ));
            }
            
            let line = Line::from(spans);
            
            // Highlight selected item
            if i + app.scroll_offset == app.selected_index {
                ListItem::new(line).style(
                    Style::default()
                        .bg(Color::DarkGray)
                        .add_modifier(Modifier::BOLD),
                )
            } else {
                ListItem::new(line)
            }
        })
        .collect();

    let activities_list = List::new(items)
        .block(
            Block::default()
                .borders(Borders::NONE)
                .title(if app.is_filtering {
                    format!("Search: {}", app.filter_text)
                } else {
                    String::new()
                }),
        );

    f.render_widget(activities_list, area);
    
    // Show scrollbar indicator if needed
    if app.filtered_activities.len() > app.viewport_height {
        draw_scrollbar(f, area, app);
    }
}

fn draw_scrollbar(f: &mut Frame, area: Rect, app: &App) {
    let scrollbar_area = Rect {
        x: area.x + area.width - 1,
        y: area.y,
        width: 1,
        height: area.height,
    };
    
    let total_items = app.filtered_activities.len();
    let viewport_height = app.viewport_height;
    
    if total_items > 0 && viewport_height > 0 {
        let scrollbar_height = (viewport_height * area.height as usize / total_items).max(1) as u16;
        let scrollbar_position = (app.scroll_offset * area.height as usize / total_items) as u16;
        
        let scrollbar = Paragraph::new("█".repeat(scrollbar_height as usize))
            .style(Style::default().fg(Color::DarkGray));
        
        let scrollbar_rect = Rect {
            x: scrollbar_area.x,
            y: scrollbar_area.y + scrollbar_position,
            width: 1,
            height: scrollbar_height.min(area.height),
        };
        
        f.render_widget(scrollbar, scrollbar_rect);
    }
}

fn draw_footer(f: &mut Frame, area: Rect, app: &App) {
    let keybinds = if app.is_filtering {
        vec![
            ("Enter", "apply"),
            ("Esc", "cancel"),
            ("Backspace", "delete"),
        ]
    } else if app.show_help {
        vec![
            ("?", "close help"),
            ("q", "quit"),
        ]
    } else {
        vec![
            ("q", "quit"),
            ("f", "filter"),
            ("/", "search"),
            ("↑↓", "nav"),
            ("space", "details"),
            ("?", "help"),
        ]
    };
    
    let keybind_text: Vec<Span> = keybinds
        .iter()
        .flat_map(|(key, desc)| {
            vec![
                Span::styled(
                    format!("[{}]", key),
                    Style::default().fg(Color::Yellow).add_modifier(Modifier::BOLD),
                ),
                Span::styled(
                    format!("{} ", desc),
                    Style::default().fg(Color::Gray),
                ),
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
    
    f.render_widget(footer, area);
}

fn draw_help(f: &mut Frame, area: Rect) {
    let help_text = vec![
        Line::from(""),
        Line::from(vec![
            Span::styled("Navigation", Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD)),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("↑/k", Style::default().fg(Color::Yellow)),
            Span::raw("         Move up"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("↓/j", Style::default().fg(Color::Yellow)),
            Span::raw("         Move down"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("PgUp/u", Style::default().fg(Color::Yellow)),
            Span::raw("      Page up"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("PgDn/d", Style::default().fg(Color::Yellow)),
            Span::raw("      Page down"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("Home/g", Style::default().fg(Color::Yellow)),
            Span::raw("      Go to top"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("End/G", Style::default().fg(Color::Yellow)),
            Span::raw("       Go to bottom"),
        ]),
        Line::from(""),
        Line::from(vec![
            Span::styled("Filtering", Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD)),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("f", Style::default().fg(Color::Yellow)),
            Span::raw("           Cycle filter mode"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("/", Style::default().fg(Color::Yellow)),
            Span::raw("           Search"),
        ]),
        Line::from(""),
        Line::from(vec![
            Span::styled("View Options", Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD)),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("Space", Style::default().fg(Color::Yellow)),
            Span::raw("       Toggle details"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("a", Style::default().fg(Color::Yellow)),
            Span::raw("           Toggle auto-scroll"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("t", Style::default().fg(Color::Yellow)),
            Span::raw("           Toggle timestamps"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("c", Style::default().fg(Color::Yellow)),
            Span::raw("           Clear activities"),
        ]),
        Line::from(""),
        Line::from(vec![
            Span::styled("General", Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD)),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("?", Style::default().fg(Color::Yellow)),
            Span::raw("           Toggle this help"),
        ]),
        Line::from(vec![
            Span::raw("  "),
            Span::styled("q", Style::default().fg(Color::Yellow)),
            Span::raw("           Quit"),
        ]),
    ];
    
    let help = Paragraph::new(help_text)
        .block(
            Block::default()
                .title(" Help ")
                .borders(Borders::ALL)
                .border_style(Style::default().fg(Color::Cyan)),
        )
        .alignment(Alignment::Left);
    
    // Center the help dialog
    let help_area = centered_rect(60, 80, area);
    f.render_widget(help, help_area);
}

fn format_filter_mode(mode: &FilterMode) -> String {
    match mode {
        FilterMode::All => "All".to_string(),
        FilterMode::Commands => "Commands".to_string(),
        FilterMode::Memory => "Memory".to_string(),
        FilterMode::FileAccess => "File Access".to_string(),
        FilterMode::ToolUsage => "Tool Usage".to_string(),
        FilterMode::Search(query) => format!("Search: {}", query),
    }
}

/// Helper function to create a centered rect
fn centered_rect(percent_x: u16, percent_y: u16, r: Rect) -> Rect {
    let popup_layout = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage((100 - percent_y) / 2),
            Constraint::Percentage(percent_y),
            Constraint::Percentage((100 - percent_y) / 2),
        ])
        .split(r);

    Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage((100 - percent_x) / 2),
            Constraint::Percentage(percent_x),
            Constraint::Percentage((100 - percent_x) / 2),
        ])
        .split(popup_layout[1])[1]
}