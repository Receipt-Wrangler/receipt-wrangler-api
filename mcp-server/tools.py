"""MCP tools for Receipt Wrangler API."""

from typing import Any, Dict, List, Optional
from mcp.server import Server
from mcp.types import Tool, TextContent
import json
from api_client import ReceiptWranglerClient


class ReceiptWranglerTools:
    def __init__(self, client: ReceiptWranglerClient, default_group_id: str = None):
        self.client = client
        self.default_group_id = default_group_id
        
    def register_tools(self, server: Server):
        """Register all tools with the MCP server."""
        
        # Receipt tools
        @server.call_tool()
        async def get_receipts(
            group_id: Optional[str] = None,
            page: Optional[int] = None,
            page_size: Optional[int] = None,
            order_by: Optional[str] = None,
            sort_direction: Optional[str] = None,
            category_ids: Optional[str] = None,
            tag_ids: Optional[str] = None,
            date_filter: Optional[str] = None,
            date_from: Optional[str] = None,
            date_to: Optional[str] = None
        ) -> List[TextContent]:
            """Get receipts with optional filtering and pagination."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                params = {}
                if page is not None:
                    params["page"] = page
                if page_size is not None:
                    params["pageSize"] = page_size
                if order_by:
                    params["orderBy"] = order_by
                if sort_direction:
                    params["sortDirection"] = sort_direction
                if category_ids:
                    params["categoryIds"] = category_ids
                if tag_ids:
                    params["tagIds"] = tag_ids
                if date_filter:
                    params["dateFilter"] = date_filter
                if date_from:
                    params["dateFrom"] = date_from
                if date_to:
                    params["dateTo"] = date_to
                
                result = await self.client.get_receipts(group_id, params)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def get_receipt(receipt_id: str) -> List[TextContent]:
            """Get a specific receipt by ID."""
            try:
                result = await self.client.get_receipt(receipt_id)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def create_receipt(
            name: str,
            date: str,
            amount: str,
            group_id: Optional[str] = None,
            description: Optional[str] = None,
            paid_by_user_id: Optional[str] = None,
            category_ids: Optional[str] = None,
            tag_ids: Optional[str] = None,
            items: Optional[str] = None
        ) -> List[TextContent]:
            """Create a new receipt."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                receipt_data = {
                    "name": name,
                    "date": date,
                    "amount": amount,
                    "groupId": group_id
                }
                
                if description:
                    receipt_data["description"] = description
                if paid_by_user_id:
                    receipt_data["paidByUserId"] = paid_by_user_id
                if category_ids:
                    receipt_data["categoryIds"] = category_ids.split(",")
                if tag_ids:
                    receipt_data["tagIds"] = tag_ids.split(",")
                if items:
                    receipt_data["receiptItems"] = json.loads(items)
                
                result = await self.client.create_receipt(receipt_data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def update_receipt(receipt_id: str, receipt_data: str) -> List[TextContent]:
            """Update an existing receipt. receipt_data should be JSON string."""
            try:
                data = json.loads(receipt_data)
                result = await self.client.update_receipt(receipt_id, data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def delete_receipt(receipt_id: str) -> List[TextContent]:
            """Delete a receipt."""
            try:
                await self.client.delete_receipt(receipt_id)
                return [TextContent(type="text", text="Receipt deleted successfully")]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # Category tools
        @server.call_tool()
        async def get_categories(group_id: Optional[str] = None) -> List[TextContent]:
            """Get categories for a group."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                result = await self.client.get_categories(group_id)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def create_category(
            name: str,
            group_id: Optional[str] = None,
            description: Optional[str] = None,
            color: Optional[str] = None
        ) -> List[TextContent]:
            """Create a new category."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                category_data = {
                    "name": name,
                    "groupId": group_id
                }
                
                if description:
                    category_data["description"] = description
                if color:
                    category_data["color"] = color
                
                result = await self.client.create_category(category_data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def update_category(category_id: str, category_data: str) -> List[TextContent]:
            """Update an existing category. category_data should be JSON string."""
            try:
                data = json.loads(category_data)
                result = await self.client.update_category(category_id, data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def delete_category(category_id: str) -> List[TextContent]:
            """Delete a category."""
            try:
                await self.client.delete_category(category_id)
                return [TextContent(type="text", text="Category deleted successfully")]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # Tag tools
        @server.call_tool()
        async def get_tags(group_id: Optional[str] = None) -> List[TextContent]:
            """Get tags for a group."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                result = await self.client.get_tags(group_id)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def create_tag(
            name: str,
            group_id: Optional[str] = None,
            description: Optional[str] = None,
            color: Optional[str] = None
        ) -> List[TextContent]:
            """Create a new tag."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                tag_data = {
                    "name": name,
                    "groupId": group_id
                }
                
                if description:
                    tag_data["description"] = description
                if color:
                    tag_data["color"] = color
                
                result = await self.client.create_tag(tag_data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def update_tag(tag_id: str, tag_data: str) -> List[TextContent]:
            """Update an existing tag. tag_data should be JSON string."""
            try:
                data = json.loads(tag_data)
                result = await self.client.update_tag(tag_id, data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def delete_tag(tag_id: str) -> List[TextContent]:
            """Delete a tag."""
            try:
                await self.client.delete_tag(tag_id)
                return [TextContent(type="text", text="Tag deleted successfully")]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # Group tools
        @server.call_tool()
        async def get_groups() -> List[TextContent]:
            """Get all groups for the current user."""
            try:
                result = await self.client.get_groups()
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def get_group(group_id: Optional[str] = None) -> List[TextContent]:
            """Get group information."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                result = await self.client.get_group(group_id)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # User tools
        @server.call_tool()
        async def get_users(group_id: Optional[str] = None) -> List[TextContent]:
            """Get users in a group."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                result = await self.client.get_users(group_id)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # Search tools
        @server.call_tool()
        async def search_receipts(search_criteria: str) -> List[TextContent]:
            """Search receipts. search_criteria should be JSON string with search parameters."""
            try:
                search_data = json.loads(search_criteria)
                result = await self.client.search_receipts(search_data)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # Dashboard tools
        @server.call_tool()
        async def get_dashboard(group_id: Optional[str] = None) -> List[TextContent]:
            """Get dashboard data for a group."""
            try:
                group_id = group_id or self.default_group_id
                if not group_id:
                    return [TextContent(type="text", text="Error: group_id is required")]
                
                result = await self.client.get_dashboard(group_id)
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        # System tools
        @server.call_tool()
        async def get_system_settings() -> List[TextContent]:
            """Get system settings."""
            try:
                result = await self.client.get_system_settings()
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
        
        @server.call_tool()
        async def get_app_data() -> List[TextContent]:
            """Get application data including user info, groups, categories, tags."""
            try:
                result = await self.client.get_app_data()
                return [TextContent(type="text", text=json.dumps(result, indent=2))]
            except Exception as e:
                return [TextContent(type="text", text=f"Error: {str(e)}")]
    
    def get_tool_definitions(self):
        """Get tool definitions for the MCP server."""
        return [
            Tool(
                name="get_receipts",
                description="Get receipts with optional filtering and pagination",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"},
                        "page": {"type": "integer", "description": "Page number for pagination"},
                        "page_size": {"type": "integer", "description": "Number of items per page"},
                        "order_by": {"type": "string", "description": "Field to order by"},
                        "sort_direction": {"type": "string", "description": "Sort direction (asc/desc)"},
                        "category_ids": {"type": "string", "description": "Comma-separated category IDs"},
                        "tag_ids": {"type": "string", "description": "Comma-separated tag IDs"},
                        "date_filter": {"type": "string", "description": "Date filter type"},
                        "date_from": {"type": "string", "description": "Start date (YYYY-MM-DD)"},
                        "date_to": {"type": "string", "description": "End date (YYYY-MM-DD)"}
                    }
                }
            ),
            Tool(
                name="get_receipt",
                description="Get a specific receipt by ID",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "receipt_id": {"type": "string", "description": "Receipt ID"}
                    },
                    "required": ["receipt_id"]
                }
            ),
            Tool(
                name="create_receipt",
                description="Create a new receipt",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "name": {"type": "string", "description": "Receipt name/description"},
                        "date": {"type": "string", "description": "Receipt date (YYYY-MM-DD)"},
                        "amount": {"type": "string", "description": "Receipt amount"},
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"},
                        "description": {"type": "string", "description": "Additional description"},
                        "paid_by_user_id": {"type": "string", "description": "User ID who paid"},
                        "category_ids": {"type": "string", "description": "Comma-separated category IDs"},
                        "tag_ids": {"type": "string", "description": "Comma-separated tag IDs"},
                        "items": {"type": "string", "description": "JSON string of receipt items"}
                    },
                    "required": ["name", "date", "amount"]
                }
            ),
            Tool(
                name="update_receipt",
                description="Update an existing receipt",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "receipt_id": {"type": "string", "description": "Receipt ID"},
                        "receipt_data": {"type": "string", "description": "JSON string of receipt data to update"}
                    },
                    "required": ["receipt_id", "receipt_data"]
                }
            ),
            Tool(
                name="delete_receipt",
                description="Delete a receipt",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "receipt_id": {"type": "string", "description": "Receipt ID"}
                    },
                    "required": ["receipt_id"]
                }
            ),
            Tool(
                name="get_categories",
                description="Get categories for a group",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"}
                    }
                }
            ),
            Tool(
                name="create_category",
                description="Create a new category",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "name": {"type": "string", "description": "Category name"},
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"},
                        "description": {"type": "string", "description": "Category description"},
                        "color": {"type": "string", "description": "Category color"}
                    },
                    "required": ["name"]
                }
            ),
            Tool(
                name="update_category",
                description="Update an existing category",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "category_id": {"type": "string", "description": "Category ID"},
                        "category_data": {"type": "string", "description": "JSON string of category data to update"}
                    },
                    "required": ["category_id", "category_data"]
                }
            ),
            Tool(
                name="delete_category",
                description="Delete a category",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "category_id": {"type": "string", "description": "Category ID"}
                    },
                    "required": ["category_id"]
                }
            ),
            Tool(
                name="get_tags",
                description="Get tags for a group",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"}
                    }
                }
            ),
            Tool(
                name="create_tag",
                description="Create a new tag",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "name": {"type": "string", "description": "Tag name"},
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"},
                        "description": {"type": "string", "description": "Tag description"},
                        "color": {"type": "string", "description": "Tag color"}
                    },
                    "required": ["name"]
                }
            ),
            Tool(
                name="update_tag",
                description="Update an existing tag",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "tag_id": {"type": "string", "description": "Tag ID"},
                        "tag_data": {"type": "string", "description": "JSON string of tag data to update"}
                    },
                    "required": ["tag_id", "tag_data"]
                }
            ),
            Tool(
                name="delete_tag",
                description="Delete a tag",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "tag_id": {"type": "string", "description": "Tag ID"}
                    },
                    "required": ["tag_id"]
                }
            ),
            Tool(
                name="get_groups",
                description="Get all groups for the current user",
                inputSchema={"type": "object", "properties": {}}
            ),
            Tool(
                name="get_group",
                description="Get group information",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"}
                    }
                }
            ),
            Tool(
                name="get_users",
                description="Get users in a group",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"}
                    }
                }
            ),
            Tool(
                name="search_receipts",
                description="Search receipts with advanced criteria",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "search_criteria": {"type": "string", "description": "JSON string with search parameters"}
                    },
                    "required": ["search_criteria"]
                }
            ),
            Tool(
                name="get_dashboard",
                description="Get dashboard data for a group",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "group_id": {"type": "string", "description": "Group ID (optional if default set)"}
                    }
                }
            ),
            Tool(
                name="get_system_settings",
                description="Get system settings",
                inputSchema={"type": "object", "properties": {}}
            ),
            Tool(
                name="get_app_data",
                description="Get application data including user info, groups, categories, tags",
                inputSchema={"type": "object", "properties": {}}
            )
        ]