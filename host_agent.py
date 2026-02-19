import asyncio
import os
import operator
from typing import Annotated, TypedDict, Literal
from dotenv import load_dotenv

from langgraph.graph import StateGraph, END
from langgraph.checkpoint.memory import MemorySaver # For the memory bank
from langchain_openai import ChatOpenAI
from langchain_core.messages import BaseMessage, HumanMessage, ToolMessage
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

load_dotenv()

class AgentState(TypedDict):
    messages: Annotated[list[BaseMessage], operator.add]

server_params = StdioServerParameters(command=r"C:\Users\anshu\code\tardigo\mcp-server.exe", args=[], env=None)

async def main():
    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            
            # 1. Setup Tools & LLM
            mcp_tools = await session.list_tools()
            llm_tools = [{"type": "function", "function": {"name": t.name, "description": t.description, "parameters": t.inputSchema}} for t in mcp_tools.tools]
            llm = ChatOpenAI(model="gpt-4o").bind_tools(llm_tools)

            # 2. Define Nodes
            def call_model(state: AgentState):
                return {"messages": [llm.invoke(state["messages"])]}

            async def call_tools(state: AgentState):
                last_message = state["messages"][-1]
                results = []
                for tool_call in last_message.tool_calls:
                    result = await session.call_tool(tool_call["name"], tool_call["args"])
                    results.append(ToolMessage(tool_call_id=tool_call["id"], content=result.content[0].text))
                return {"messages": results}

            def should_continue(state: AgentState) -> Literal["tools", END]:
                if state["messages"][-1].tool_calls: return "tools"
                return END

            # 3. Build Graph with Memory
            workflow = StateGraph(AgentState)
            workflow.add_node("agent", call_model)
            workflow.add_node("tools", call_tools)
            workflow.set_entry_point("agent")
            workflow.add_conditional_edges("agent", should_continue)
            workflow.add_edge("tools", "agent")
            
            # MemorySaver keeps your GRE context alive!
            memory = MemorySaver()
            app = workflow.compile(checkpointer=memory)

            # 4. THE INTERACTIVE LOOP
            print("--- TardiGo Agent Connected ---")
            print("Type 'exit' or 'quit' to stop.")
            
            # config is required for memory to work
            config = {"configurable": {"thread_id": "GRE_PREP_SESSION_1"}}

            while True:
                user_input = input("\nYou: ")
                if user_input.lower() in ["exit", "quit"]:
                    break

                inputs = {"messages": [HumanMessage(content=user_input)]}
                
                async for output in app.astream(inputs, config, stream_mode="updates"):
                    for node, data in output.items():
                        if node == "agent" and data["messages"][-1].content:
                            print(f"\nTardiGo: {data['messages'][-1].content}")

if __name__ == "__main__":
    asyncio.run(main())