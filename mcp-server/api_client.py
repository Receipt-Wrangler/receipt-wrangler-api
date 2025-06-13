"""API client wrapper for Receipt Wrangler API calls."""

import httpx
from typing import Optional, Dict, Any, List, Union
import json
from auth import AuthManager


class ReceiptWranglerClient:
    def __init__(self, auth_manager: AuthManager):
        self.auth = auth_manager
        self.base_url = auth_manager.base_url
        
    async def _make_request(
        self, 
        method: str, 
        endpoint: str, 
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
        timeout: float = 30.0
    ) -> httpx.Response:
        """Make an authenticated HTTP request."""
        if not await self.auth.ensure_valid_token():
            raise Exception("Authentication required")
            
        url = f"{self.base_url}{endpoint}"
        headers = self.auth.get_auth_headers()
        headers["Content-Type"] = "application/json"
        
        async with httpx.AsyncClient() as client:
            if method.upper() == "GET":
                response = await client.get(url, headers=headers, params=params, timeout=timeout)
            elif method.upper() == "POST":
                response = await client.post(url, headers=headers, json=data, params=params, timeout=timeout)
            elif method.upper() == "PUT":
                response = await client.put(url, headers=headers, json=data, params=params, timeout=timeout)
            elif method.upper() == "DELETE":
                response = await client.delete(url, headers=headers, params=params, timeout=timeout)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
                
        return response
    
    # Receipt operations
    async def get_receipts(self, group_id: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Get receipts for a group with optional filtering."""
        response = await self._make_request("GET", f"/receipts/group/{group_id}", params=params)
        response.raise_for_status()
        return response.json()
    
    async def get_receipt(self, receipt_id: str) -> Dict[str, Any]:
        """Get a specific receipt by ID."""
        response = await self._make_request("GET", f"/receipts/{receipt_id}")
        response.raise_for_status()
        return response.json()
    
    async def create_receipt(self, receipt_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new receipt."""
        response = await self._make_request("POST", "/receipts/", data=receipt_data)
        response.raise_for_status()
        return response.json()
    
    async def update_receipt(self, receipt_id: str, receipt_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update an existing receipt."""
        response = await self._make_request("PUT", f"/receipts/{receipt_id}", data=receipt_data)
        response.raise_for_status()
        return response.json()
    
    async def delete_receipt(self, receipt_id: str) -> bool:
        """Delete a receipt."""
        response = await self._make_request("DELETE", f"/receipts/{receipt_id}")
        response.raise_for_status()
        return True
    
    # Category operations
    async def get_categories(self, group_id: str) -> List[Dict[str, Any]]:
        """Get categories for a group."""
        response = await self._make_request("GET", f"/categories/group/{group_id}")
        response.raise_for_status()
        return response.json()
    
    async def create_category(self, category_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new category."""
        response = await self._make_request("POST", "/categories/", data=category_data)
        response.raise_for_status()
        return response.json()
    
    async def update_category(self, category_id: str, category_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update an existing category."""
        response = await self._make_request("PUT", f"/categories/{category_id}", data=category_data)
        response.raise_for_status()
        return response.json()
    
    async def delete_category(self, category_id: str) -> bool:
        """Delete a category."""
        response = await self._make_request("DELETE", f"/categories/{category_id}")
        response.raise_for_status()
        return True
    
    # Tag operations
    async def get_tags(self, group_id: str) -> List[Dict[str, Any]]:
        """Get tags for a group."""
        response = await self._make_request("GET", f"/tags/group/{group_id}")
        response.raise_for_status()
        return response.json()
    
    async def create_tag(self, tag_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new tag."""
        response = await self._make_request("POST", "/tags/", data=tag_data)
        response.raise_for_status()
        return response.json()
    
    async def update_tag(self, tag_id: str, tag_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update an existing tag."""
        response = await self._make_request("PUT", f"/tags/{tag_id}", data=tag_data)
        response.raise_for_status()
        return response.json()
    
    async def delete_tag(self, tag_id: str) -> bool:
        """Delete a tag."""
        response = await self._make_request("DELETE", f"/tags/{tag_id}")
        response.raise_for_status()
        return True
    
    # Group operations
    async def get_group(self, group_id: str) -> Dict[str, Any]:
        """Get group information."""
        response = await self._make_request("GET", f"/group/{group_id}")
        response.raise_for_status()
        return response.json()
    
    async def get_groups(self) -> List[Dict[str, Any]]:
        """Get all groups for the current user."""
        response = await self._make_request("GET", "/groups/")
        response.raise_for_status()
        return response.json()
    
    async def create_group(self, group_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new group."""
        response = await self._make_request("POST", "/group/", data=group_data)
        response.raise_for_status()
        return response.json()
    
    async def update_group(self, group_id: str, group_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update an existing group."""
        response = await self._make_request("PUT", f"/group/{group_id}", data=group_data)
        response.raise_for_status()
        return response.json()
    
    # User operations
    async def get_users(self, group_id: str) -> List[Dict[str, Any]]:
        """Get users in a group."""
        response = await self._make_request("GET", f"/users/group/{group_id}")
        response.raise_for_status()
        return response.json()
    
    async def get_user(self, user_id: str) -> Dict[str, Any]:
        """Get user information."""
        response = await self._make_request("GET", f"/users/{user_id}")
        response.raise_for_status()
        return response.json()
    
    # Search operations
    async def search_receipts(self, search_data: Dict[str, Any]) -> Dict[str, Any]:
        """Search receipts with filters."""
        response = await self._make_request("POST", "/search/", data=search_data)
        response.raise_for_status()
        return response.json()
    
    # Dashboard operations
    async def get_dashboard(self, group_id: str) -> List[Dict[str, Any]]:
        """Get dashboard data for a group."""
        response = await self._make_request("GET", f"/dashboard/{group_id}")
        response.raise_for_status()
        return response.json()
    
    # System operations
    async def get_system_settings(self) -> Dict[str, Any]:
        """Get system settings."""
        response = await self._make_request("GET", "/systemSettings/")
        response.raise_for_status()
        return response.json()
    
    async def get_app_data(self) -> Dict[str, Any]:
        """Get application data."""
        response = await self._make_request("GET", "/appData/")
        response.raise_for_status()
        return response.json()