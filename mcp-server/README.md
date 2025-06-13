# Receipt Wrangler MCP Server

This is an MCP (Model Context Protocol) server that provides full API access to Receipt Wrangler through Claude.

## Setup

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Configure your Receipt Wrangler instance:
   - Set environment variables (optional):
     ```bash
     export RECEIPT_WRANGLER_URL="http://localhost:8081"
     export RECEIPT_WRANGLER_USERNAME="your_username"
     export RECEIPT_WRANGLER_PASSWORD="your_password"
     export RECEIPT_WRANGLER_GROUP_ID="your_default_group_id"
     ```
   - Or the server will prompt you for these values at startup

3. Run the server:
```bash
python server.py
```

## Available Tools

The MCP server provides the following tools:

### Receipt Management
- `get_receipts` - Get receipts with filtering and pagination
- `get_receipt` - Get a specific receipt by ID
- `create_receipt` - Create a new receipt
- `update_receipt` - Update an existing receipt
- `delete_receipt` - Delete a receipt

### Categories
- `get_categories` - Get categories for a group
- `create_category` - Create a new category
- `update_category` - Update an existing category
- `delete_category` - Delete a category

### Tags
- `get_tags` - Get tags for a group
- `create_tag` - Create a new tag
- `update_tag` - Update an existing tag
- `delete_tag` - Delete a tag

### Groups & Users
- `get_groups` - Get all groups for the current user
- `get_group` - Get group information
- `get_users` - Get users in a group

### Search & Analytics
- `search_receipts` - Advanced receipt search
- `get_dashboard` - Get dashboard data for analytics

### System
- `get_system_settings` - Get system settings
- `get_app_data` - Get application data including user info, groups, categories, tags

## Usage with Claude

Once the server is running, you can use it with Claude to manage your receipts through natural language:

- "Show me all receipts from last month"
- "Create a new category called 'Office Supplies'"
- "Add a receipt for $25.99 for groceries today"
- "Search for all receipts tagged with 'business'"
- "Show me spending analytics for this group"

## Authentication

The server handles authentication automatically:
- Prompts for credentials at startup if not provided via environment variables
- Automatically refreshes tokens as needed
- Validates group access before operations

## Error Handling

All tools include comprehensive error handling and will return descriptive error messages if operations fail.