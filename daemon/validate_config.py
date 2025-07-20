#!/usr/bin/env python3
"""Validate agents.json configuration"""

import json
import sys

def validate_config():
    try:
        with open('agents.json', 'r') as f:
            config = json.load(f)
        
        # Check required top-level keys
        required_keys = ['base_guidance', 'agents', 'model_config', 'response_config']
        for key in required_keys:
            if key not in config:
                print(f"❌ Missing required key: {key}")
                return False
        
        # Validate agents
        for agent_id, agent in config['agents'].items():
            print(f"✅ Agent '{agent_id}': {agent['name']}")
            if 'prompt' not in agent:
                print(f"  ⚠️  Missing prompt for {agent_id}")
        
        # Validate model config
        if 'default' not in config['model_config']:
            print("❌ Missing default model in model_config")
            return False
        
        print(f"\n✅ Configuration valid!")
        print(f"   Default model: {config['model_config']['default']}")
        print(f"   Agents configured: {len(config['agents'])}")
        return True
        
    except json.JSONDecodeError as e:
        print(f"❌ Invalid JSON: {e}")
        return False
    except FileNotFoundError:
        print("❌ agents.json not found")
        return False
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

if __name__ == "__main__":
    sys.exit(0 if validate_config() else 1)