#!/usr/bin/env python3
"""
Test Function Calling Support for Kolosal.ai and OpenRouter
Usage: python3 backend/testing/function_calling_test.py
"""

import os
import json
import requests
import sys
from typing import Dict, List, Any, Optional

def test_kolosal(api_key: str, tools: List[Dict], user_message: str):
    """Test Kolosal.ai API with function calling"""
    print("📡 Testing Kolosal.ai API (GLM 4.6)")
    print("-----------------------------------")
    
    url = "https://api.kolosal.ai/v1/chat/completions"
    
    payload = {
        "model": "GLM 4.6",
        "messages": [
            {
                "role": "system",
                "content": "Kamu adalah Asisten Bantuaku. Gunakan tools yang tersedia untuk menyimpan informasi perusahaan dan produk."
            },
            {
                "role": "user",
                "content": user_message
            }
        ],
        "tools": tools,
        "tool_choice": "auto",
        "max_tokens": 2000,
        "temperature": 0.7
    }
    
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}"
    }
    
    print(f"📤 Request URL: {url}")
    print(f"📤 Model: {payload['model']}")
    print(f"📤 Tool Choice: {payload['tool_choice']}")
    print(f"📤 Tools Count: {len(payload['tools'])}")
    print(f"📤 Request Body (first 500 chars): {json.dumps(payload, indent=2)[:500]}...\n")
    
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=60)
        
        print(f"📥 Response Status: {response.status_code} {response.reason}")
        
        if response.status_code != 200:
            print(f"❌ API Error Response:\n{response.text}")
            return
        
        data = response.json()
        
        if "choices" not in data or len(data["choices"]) == 0:
            print("❌ No choices in response")
            print(f"Raw response: {json.dumps(data, indent=2)}")
            return
        
        choice = data["choices"][0]
        message = choice.get("message", {})
        
        print(f"📥 Finish Reason: {choice.get('finish_reason', 'N/A')}")
        print(f"📥 Message Role: {message.get('role', 'N/A')}")
        print(f"📥 Message Content: {message.get('content', 'N/A')}")
        
        tool_calls = message.get("tool_calls", [])
        if tool_calls:
            print("✅ FUNCTION CALLING SUPPORTED!")
            print(f"📥 Tool Calls Count: {len(tool_calls)}")
            for i, tool_call in enumerate(tool_calls):
                print(f"\n  Tool Call #{i+1}:")
                print(f"    ID: {tool_call.get('id', 'N/A')}")
                print(f"    Type: {tool_call.get('type', 'N/A')}")
                func = tool_call.get("function", {})
                print(f"    Function Name: {func.get('name', 'N/A')}")
                print(f"    Function Arguments: {func.get('arguments', 'N/A')}")
        else:
            print("⚠️  No tool calls in response")
            print("   This could mean:")
            print("   1. Function calling not supported by Kolosal.ai API")
            print("   2. Model chose not to use tools (check finish_reason)")
            print("   3. API ignored tools parameter")
        
        if "usage" in data:
            usage = data["usage"]
            print("\n📊 Token Usage:")
            print(f"   Prompt: {usage.get('prompt_tokens', 0)}")
            print(f"   Completion: {usage.get('completion_tokens', 0)}")
            print(f"   Total: {usage.get('total_tokens', 0)}")
    
    except requests.exceptions.RequestException as e:
        print(f"❌ Request failed: {e}")
    except json.JSONDecodeError as e:
        print(f"❌ Failed to parse response: {e}")
        print(f"Raw response: {response.text}")


def test_openrouter(api_key: str, tools: List[Dict], user_message: str):
    """Test OpenRouter API with function calling"""
    print("📡 Testing OpenRouter API (x-ai/grok-4-fast)")
    print("--------------------------------------------")
    
    url = "https://openrouter.ai/api/v1/chat/completions"
    
    payload = {
        "model": "x-ai/grok-4-fast",
        "messages": [
            {
                "role": "system",
                "content": "You are Bantuaku Assistant. Use available tools to store company and product information."
            },
            {
                "role": "user",
                "content": user_message
            }
        ],
        "tools": tools,
        "tool_choice": "auto",
        "max_tokens": 2000,
        "temperature": 0.7
    }
    
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}",
        "HTTP-Referer": "https://bantuaku.com",
        "X-Title": "Bantuaku Function Calling Test"
    }
    
    print(f"📤 Request URL: {url}")
    print(f"📤 Model: {payload['model']}")
    print(f"📤 Tool Choice: {payload['tool_choice']}")
    print(f"📤 Tools Count: {len(payload['tools'])}")
    print(f"📤 Request Body (first 500 chars): {json.dumps(payload, indent=2)[:500]}...\n")
    
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=60)
        
        print(f"📥 Response Status: {response.status_code} {response.reason}")
        
        if response.status_code != 200:
            print(f"❌ API Error Response:\n{response.text}")
            return
        
        data = response.json()
        
        if "choices" not in data or len(data["choices"]) == 0:
            print("❌ No choices in response")
            print(f"Raw response: {json.dumps(data, indent=2)}")
            return
        
        choice = data["choices"][0]
        message = choice.get("message", {})
        
        print(f"📥 Finish Reason: {choice.get('finish_reason', 'N/A')}")
        print(f"📥 Message Role: {message.get('role', 'N/A')}")
        print(f"📥 Message Content: {message.get('content', 'N/A')}")
        
        tool_calls = message.get("tool_calls", [])
        if tool_calls:
            print("✅ FUNCTION CALLING SUPPORTED!")
            print(f"📥 Tool Calls Count: {len(tool_calls)}")
            for i, tool_call in enumerate(tool_calls):
                print(f"\n  Tool Call #{i+1}:")
                print(f"    ID: {tool_call.get('id', 'N/A')}")
                print(f"    Type: {tool_call.get('type', 'N/A')}")
                func = tool_call.get("function", {})
                print(f"    Function Name: {func.get('name', 'N/A')}")
                print(f"    Function Arguments: {func.get('arguments', 'N/A')}")
        else:
            print("⚠️  No tool calls in response")
            print("   This could mean:")
            print("   1. Function calling not supported by model")
            print("   2. Model chose not to use tools (check finish_reason)")
            print("   3. API ignored tools parameter")
        
        if "usage" in data:
            usage = data["usage"]
            print("\n📊 Token Usage:")
            print(f"   Prompt: {usage.get('prompt_tokens', 0)}")
            print(f"   Completion: {usage.get('completion_tokens', 0)}")
            print(f"   Total: {usage.get('total_tokens', 0)}")
    
    except requests.exceptions.RequestException as e:
        print(f"❌ Request failed: {e}")
    except json.JSONDecodeError as e:
        print(f"❌ Failed to parse response: {e}")
        print(f"Raw response: {response.text}")


def main():
    print("🧪 Testing Function Calling Support")
    print("=====================================\n")
    
    # Define test tools
    tools = [
        {
            "type": "function",
            "function": {
                "name": "update_company_info",
                "description": "Update company information fields (industry, location, business model)",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "industry": {
                            "type": "string",
                            "description": "Business industry/type (e.g., 'Kuliner', 'Retail', 'Jasa')"
                        },
                        "city": {
                            "type": "string",
                            "description": "City where business operates"
                        },
                        "location_region": {
                            "type": "string",
                            "description": "Region/province"
                        }
                    },
                    "required": ["industry", "city"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "create_product",
                "description": "Create a new product or service for the company",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "name": {
                            "type": "string",
                            "description": "Product/service name"
                        },
                        "category": {
                            "type": "string",
                            "description": "Product category"
                        },
                        "unit_price": {
                            "type": "number",
                            "description": "Price per unit"
                        }
                    },
                    "required": ["name"]
                }
            }
        }
    ]
    
    test_message = "Saya punya bisnis kuliner di Jakarta. Produk utama saya adalah Nasi Goreng Spesial dengan harga 25000."
    
    # Test Kolosal.ai
    kolosal_key = os.getenv("KOLOSAL_API_KEY")
    if kolosal_key:
        test_kolosal(kolosal_key, tools, test_message)
    else:
        print("❌ KOLOSAL_API_KEY not set, skipping Kolosal test")
    
    print("\n")
    
    # Test OpenRouter
    openrouter_key = os.getenv("OPENROUTER_API_KEY")
    if openrouter_key:
        test_openrouter(openrouter_key, tools, test_message)
    else:
        print("❌ OPENROUTER_API_KEY not set, skipping OpenRouter test")
    
    print("\n✅ Testing complete!")


if __name__ == "__main__":
    main()
