#!/usr/bin/env node
// mark-mcp 入口：独立进程，通过 stdio 与 Claude Code 通信

import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { createMcpServer } from "./mcp/server.js";

const server = createMcpServer();
const transport = new StdioServerTransport();
await server.connect(transport);
