#!/usr/bin/env python3
"""Basic test script for Receipt Wrangler MCP Server components."""

import asyncio
import sys
import os

# Add current directory to path for imports
sys.path.insert(0, os.path.dirname(__file__))

from config import Config
from auth import AuthManager


async def test_basic_functionality():
    """Test basic functionality without actual API calls."""
    print("Testing basic functionality...")
    
    # Test config
    config = Config()
    config.base_url = "http://localhost:8081/api"
    config.username = "test_user"
    config.password = "test_pass"
    print("✓ Config creation works")
    
    # Test auth manager creation
    auth_manager = AuthManager(config.base_url)
    print("✓ AuthManager creation works")
    
    # Test auth headers (without token)
    headers = auth_manager.get_auth_headers()
    assert headers == {}, "Should return empty headers without token"
    print("✓ Auth headers work correctly")
    
    print("\nBasic functionality tests passed!")
    print("To test with a real Receipt Wrangler instance, run:")
    print("  python server.py")


if __name__ == "__main__":
    asyncio.run(test_basic_functionality())