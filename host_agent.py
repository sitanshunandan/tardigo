import asyncio
import os
from dotenv import load_dotenv
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, ToolMessage

load_dotenv() # Loads your OPENAI_API_KEY from .env

server_params = StdioServerParameters(
    command=r"C:\Users\anshu\code\tardigo\mcp-server.exe",
    args=[],
    env=None
)

async def run_reasoning_agent():
    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            
            # 1. Fetch tool definitions from your Go Server
            mcp_tools = await session.list_tools()
            
            # 2. Convert MCP tools to a format OpenAI understands
            # We map the Go tool definition to the LLM's "tools" parameter
            llm_tools = [
                {
                    "type": "function",
                    "function": {
                        "name": t.name,
                        "description": t.description,
                        "parameters": t.inputSchema,
                    },
                }
                for t in mcp_tools.tools
            ]

            llm = ChatOpenAI(model="gpt-4o") # Use gpt-4o for best reasoning
            
            # 3. The Conversation
            #user_input = "I woke up at 7:30 AM. I need to spend 90 minutes on a 'Difficult Kernel Bug' (level 9) and 30 minutes on 'Email' (level 2). When should I do them?"
            user_input = """
                I woke up at 8:00 AM. I have a very busy day:
                1. 'Refactor Go Concurrency' (90 mins, Level 10) - Very hard.
                2. 'Weekly Sync' (30 mins, Level 3) - Easy.
                3. 'Write Documentation' (60 mins, Level 5) - Medium.
                Please plan these to maximize my brain power.
            """            
            print(f"\nUser: {user_input}")

            # Step A: LLM decides which tool to call
            messages = [HumanMessage(content=user_input)]
            response = llm.invoke(messages, tools=llm_tools)
            
            if response.tool_calls:
                for tool_call in response.tool_calls:
                    print(f"\nAgent is thinking... 'I should use the {tool_call['name']} tool.'")
                    
                    # Step B: Python executes the Go tool on behalf of the LLM
                    result = await session.call_tool(tool_call["name"], tool_call["args"])
                    
                    # Step C: Feed the Go output back to the LLM for a final natural language answer
                    messages.append(response)
                    messages.append(ToolMessage(content=result.content[0].text, tool_call_id=tool_call["id"]))
                    
                    final_answer = llm.invoke(messages)
                    print(f"\nAgent: {final_answer.content}")

if __name__ == "__main__":
    asyncio.run(run_reasoning_agent())