# Data Chatter Web Interface

A beautiful, responsive web interface for interacting with the Data Chatter API.

## Features

- 🎨 **Modern UI**: Clean, responsive design with gradient backgrounds
- 🔍 **Natural Language Queries**: Ask questions in plain English
- 📊 **Data Tables**: Results displayed in beautiful, sortable tables
- 📱 **Mobile Friendly**: Responsive design works on all devices
- ⚡ **Real-time**: Instant query execution and results
- 🎯 **Smart Queries**: Automatically discovers schema and executes queries

## Quick Start

### Option 1: Use the Full Stack Script
```bash
# From the project root
./scripts/start_full_stack.sh
```

### Option 2: Manual Start
```bash
# Terminal 1: Start API server
cd /Users/stephaniegredell/data-chatter
./bin/server

# Terminal 2: Start web server
cd /Users/stephaniegredell/data-chatter/web
go run server.go
```

Then open http://localhost:3000 in your browser.

## Example Queries

- "fetch me all contacts available on Monday"
- "show me all contacts from California"
- "find contacts with email addresses containing 'gmail'"
- "get all contacts created this week"

## API Integration

The web interface automatically connects to your API server running on port 8081. Make sure your API server is running before starting the web interface.

## Features

### Query Interface
- Large text area for natural language queries
- Example query buttons for quick testing
- Real-time loading indicators
- Error handling and display

### Results Display
- Responsive data tables
- Column name formatting
- Cell value truncation for long text
- Query information display
- Result count indicators

### Responsive Design
- Mobile-first approach
- Collapsible tables on small screens
- Touch-friendly interface
- Optimized for all screen sizes

## Customization

The interface uses vanilla HTML, CSS, and JavaScript - no frameworks required. You can easily customize:

- Colors and styling in the `<style>` section
- API endpoint in the JavaScript
- Table formatting functions
- Query examples

## Browser Support

- Chrome/Chromium (recommended)
- Firefox
- Safari
- Edge

Requires modern JavaScript features (ES6+).
