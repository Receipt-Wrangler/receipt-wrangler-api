"""Authentication management for Receipt Wrangler MCP Server."""

import asyncio
from typing import Optional, Dict, Any
import httpx
from datetime import datetime, timedelta
import json
import sys


class AuthManager:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.token_expires_at: Optional[datetime] = None
        self.user_groups: list = []
        
    async def login(self, username: str, password: str) -> bool:
        """Login and store tokens."""
        login_data = {
            "username": username,
            "password": password
        }
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/login/",
                    json=login_data,
                    params={"tokensInBody": "true"},  # Get tokens in response body
                    timeout=30.0
                )
                
                if response.status_code == 200:
                    app_data = response.json()
                    self.access_token = app_data.get("jwt")
                    self.refresh_token = app_data.get("refreshToken")
                    self.user_groups = app_data.get("groups", [])
                    
                    # Estimate token expiration (JWT tokens typically last 1 hour)
                    self.token_expires_at = datetime.utcnow() + timedelta(minutes=55)
                    
                    print(f"Successfully authenticated! Available groups: {len(self.user_groups)}")
                    for group in self.user_groups:
                        print(f"  - {group.get('name', 'Unknown')} (ID: {group.get('id')})")
                    
                    return True
                else:
                    print(f"Login failed: {response.status_code} - {response.text}")
                    return False
                    
        except Exception as e:
            print(f"Login error: {e}")
            return False
    
    async def refresh_access_token(self) -> bool:
        """Refresh the access token using refresh token."""
        if not self.refresh_token:
            return False
            
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/token/refresh",
                    json={"refreshToken": self.refresh_token},
                    timeout=30.0
                )
                
                if response.status_code == 200:
                    token_data = response.json()
                    self.access_token = token_data.get("jwt")
                    # Update refresh token if provided
                    if "refreshToken" in token_data:
                        self.refresh_token = token_data["refreshToken"]
                    
                    self.token_expires_at = datetime.utcnow() + timedelta(minutes=55)
                    return True
                else:
                    return False
                    
        except Exception as e:
            print(f"Token refresh error: {e}")
            return False
    
    async def ensure_valid_token(self) -> bool:
        """Ensure we have a valid access token, refresh if needed."""
        if not self.access_token:
            return False
            
        # Check if token is close to expiring (5 minute buffer)
        if self.token_expires_at and datetime.utcnow() >= (self.token_expires_at - timedelta(minutes=5)):
            print("Token expiring soon, refreshing...")
            if not await self.refresh_access_token():
                print("Token refresh failed, authentication required")
                return False
                
        return True
    
    def get_auth_headers(self) -> Dict[str, str]:
        """Get headers for authenticated requests."""
        if not self.access_token:
            return {}
        return {"Authorization": f"Bearer {self.access_token}"}
    
    async def validate_group_access(self, group_id: str) -> bool:
        """Validate that user has access to the specified group."""
        if not await self.ensure_valid_token():
            return False
            
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(
                    f"{self.base_url}/group/{group_id}",
                    headers=self.get_auth_headers(),
                    timeout=30.0
                )
                
                if response.status_code == 200:
                    return True
                elif response.status_code == 403:
                    return False
                else:
                    print(f"Error validating group access: {response.status_code}")
                    return False
                    
        except Exception as e:
            print(f"Group validation error: {e}")
            return False