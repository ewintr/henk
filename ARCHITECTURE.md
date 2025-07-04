# Henk Architecture Document

## Overview

Henk is a Go-based AI assistant CLI tool designed for collabo1rative software development. It acts as a coding partner that provides guidance and support while keeping the developer in control of all code modifications.

## Core Philosophy

- Non-invasive: Henk never directly modifies code files
- Collaborative: Functions as a pair programming partner, not a replacement 
- Tool-agnostic: Integrates with existing developer workflows
- Privacy-conscious: Supports local LLMs for complete data control

## System Architecture

### Main Components
                                                                            
  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐ 
  │    main.go  │    │   config/   │    │    agent/   │ 
  │             │───▶│  providers  │───▶│   core      │ 
  │ Entry Point │    │   models    │    │  logic      │ 
  └─────────────┘    └─────────────┘    └──────┬──────┘ 
                                               │                           
                     ┌─────────────┐    ┌──────▼──────┐                     
                     │    llm/     │    │    tool/    │                     
                     │  providers  │◀───│  system     │                     
                     └─────────────┘    └─────────────┘                     
                             │                                              
                     ┌───────▼─────────┐                                    
                     │      ui/        │                                    
                     │  interaction    │                                    
                     └─────────────────┘                                    

### Package Structure 

####  /agent  - Core Agent Logic

-  agent.go : Main agent orchestration and conversation loop
-  config.go : Configuration management and validation
-  command.go : CLI command processing
-  ui.go : User interface handling with channels

####  /agent/llm  - LLM Integration Layer

-  llm.go : Provider-agnostic LLM interface
-  claude.go : Anthropic Claude integration
-  openai.go : OpenAI API integration
-  ollama.go : Local Ollama integration

####  /agent/tool  - Tool System

-  tool.go : Tool interface definition 
-  readfile.go : File reading capability
-  listfiles.go : Directory listing capability

## Key Design Patterns

### Provider Pattern

Multiple LLM providers (Claude, OpenAI, Ollama) implement a common LLM interface, allowing seamless switching between local and remote models.

### Tool System Architecture

- Tools implement a common interface with JSON schema validation
- Tools are injected into the agent and made available to LLMs
- Currently includes file reading and directory listing tools

### Message-Based Architecture

- Asynchronous communication via Go channels
- Separate input/output channels for clean UI separation
- Conversation state maintained as message history

## Configuration System

### TOML-Based Configuration

- User config directory:  ~/.config/henk/config.toml
- Provider configurations with models and API keys
- Environment variable support for API keys
- Configurable system prompts

### Provider Configuration

Each provider supports:

- Multiple models with short names
- Custom base URLs (for self-hosted services)
- Context size limits
- Default model selection

## Data Flow

1. Initialization: Load config, validate providers, initialize LLM client
2. Conversation Loop:
  - Accept user input via UI channel
  - Process commands or send to LLM
  - Handle tool execution requests from LLM
  - Return results and continue conversation
3. Tool Execution: LLM requests tools → Agent executes → Results fed back to LLM

## Extension Points

### Adding New Tools

1. Implement the  Tool  interface in  /agent/tool/
2. Register tool in  main.go
3. Tool automatically becomes available to LLMs via JSON schema

### Adding New LLM Providers

1. Implement the  LLM  interface in  /agent/llm/
2. Add provider type to  NewLLM()  factory function
3. Configure in  config.toml

## Dependencies

- UI: Charmbracelet ecosystem (glamour, huh, bubbletea)
- LLM SDKs: Official Anthropic, OpenAI Go SDKs
- Configuration: BurntSushi/toml
- JSON Schema: invopop/jsonschema for tool definitions

## Future Considerations

- API server for editor integration
- Additional tool capabilities
- Plugin system for custom tools
- Enhanced conversation management
