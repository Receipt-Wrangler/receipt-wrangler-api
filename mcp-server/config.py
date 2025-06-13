"""Configuration management for Receipt Wrangler MCP Server."""

import os
from typing import Optional
from dotenv import load_dotenv

load_dotenv()


class Config:
    def __init__(self):
        self.base_url: Optional[str] = None
        self.username: Optional[str] = None
        self.password: Optional[str] = None
        self.group_id: Optional[str] = None
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        
    def load_from_env(self):
        """Load configuration from environment variables."""
        self.base_url = os.getenv('RECEIPT_WRANGLER_URL')
        self.username = os.getenv('RECEIPT_WRANGLER_USERNAME')
        self.password = os.getenv('RECEIPT_WRANGLER_PASSWORD')
        self.group_id = os.getenv('RECEIPT_WRANGLER_GROUP_ID')
        
    def prompt_for_config(self):
        """Prompt user for configuration if not provided via environment."""
        if not self.base_url:
            self.base_url = input("Enter Receipt Wrangler URL (e.g., http://localhost:8081): ").strip()
            if not self.base_url.startswith(('http://', 'https://')):
                self.base_url = f"http://{self.base_url}"
            if not self.base_url.endswith('/api'):
                self.base_url = f"{self.base_url}/api"
                
        if not self.username:
            self.username = input("Enter username: ").strip()
            
        if not self.password:
            import getpass
            self.password = getpass.getpass("Enter password: ").strip()


config = Config()