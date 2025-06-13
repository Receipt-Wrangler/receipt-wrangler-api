#!/usr/bin/env python3
"""Setup script for Receipt Wrangler MCP Server."""

import subprocess
import sys
import os

def main():
    """Setup the MCP server environment."""
    print("Setting up Receipt Wrangler MCP Server...")
    
    # Install dependencies
    print("Installing dependencies...")
    try:
        subprocess.run([sys.executable, "-m", "pip", "install", "-r", "requirements.txt"], 
                      check=True, cwd=os.path.dirname(__file__))
        print("Dependencies installed successfully!")
    except subprocess.CalledProcessError as e:
        print(f"Error installing dependencies: {e}")
        sys.exit(1)
    
    # Make server executable
    server_path = os.path.join(os.path.dirname(__file__), "server.py")
    os.chmod(server_path, 0o755)
    
    print("\nSetup complete!")
    print("\nTo run the server:")
    print("  python server.py")
    print("\nOr set environment variables first:")
    print("  export RECEIPT_WRANGLER_URL='http://localhost:8081'")
    print("  export RECEIPT_WRANGLER_USERNAME='your_username'")
    print("  export RECEIPT_WRANGLER_PASSWORD='your_password'")
    print("  python server.py")

if __name__ == "__main__":
    main()