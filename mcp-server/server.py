#!/usr/bin/env python3
"""Receipt Wrangler MCP Server."""

import asyncio
import sys
from mcp.server import Server
from mcp.server.stdio import stdio_server
from config import config
from auth import AuthManager
from api_client import ReceiptWranglerClient
from tools import ReceiptWranglerTools


async def main():
    """Main server entry point."""
    # Load configuration
    config.load_from_env()
    config.prompt_for_config()
    
    # Initialize authentication
    auth_manager = AuthManager(config.base_url)
    
    print("Authenticating with Receipt Wrangler...")
    if not await auth_manager.login(config.username, config.password):
        print("Authentication failed. Exiting.")
        sys.exit(1)
    
    # Prompt for group ID if not provided
    if not config.group_id and auth_manager.user_groups:
        print("\nAvailable groups:")
        for i, group in enumerate(auth_manager.user_groups):
            print(f"{i + 1}. {group.get('name', 'Unknown')} (ID: {group.get('id')})")
        
        while True:
            try:
                choice = input("\nSelect a group (enter number or group ID): ").strip()
                # Try to parse as number first
                try:
                    choice_num = int(choice) - 1
                    if 0 <= choice_num < len(auth_manager.user_groups):
                        config.group_id = auth_manager.user_groups[choice_num]['id']
                        break
                except ValueError:
                    # Try as direct group ID
                    if await auth_manager.validate_group_access(choice):
                        config.group_id = choice
                        break
                    else:
                        print("Invalid selection or no access to that group. Please try again.")
            except (ValueError, KeyboardInterrupt):
                print("Invalid selection. Please try again.")
    
    # Initialize client and tools
    client = ReceiptWranglerClient(auth_manager)
    tools = ReceiptWranglerTools(client, config.group_id)
    
    # Create MCP server
    server = Server("receipt-wrangler")
    
    # Register tools
    tools.register_tools(server)
    
    # The tools are automatically registered by the @server.call_tool() decorators
    
    print(f"Receipt Wrangler MCP Server starting...")
    print(f"Base URL: {config.base_url}")
    print(f"Default Group ID: {config.group_id}")
    print(f"Available tools: {len(tools.get_tool_definitions())}")
    
    # Run the server
    async with stdio_server() as streams:
        await server.run(streams[0], streams[1], server.create_initialization_options())


if __name__ == "__main__":
    asyncio.run(main())