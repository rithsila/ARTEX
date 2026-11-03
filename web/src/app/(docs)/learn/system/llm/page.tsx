"use client";

import { Brain } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function LLMDocPage() {
  return (
    <DocModuleTemplate
      title="LLM"
      description="Configure upstream AI model providers, API credentials, base URLs, and token generation parameters."
      groupLabel="SYSTEM"
      icon={Brain}
      consoleRoute="/system/llm"
      badge="AI Providers"
      purpose="Manages authentication and connectivity to Large Language Model providers. ARTEX supports OpenAI, Anthropic, and any OpenAI-compatible API endpoint (such as Ollama, vLLM, DeepSeek, or LM Studio)."
      prerequisites={[
        "An API key or active local URL for your preferred model provider.",
        "System administrator access to ARTEX.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to LLM Settings",
          detail: "Click 'LLM' in the left sidebar under SYSTEM or browse to /system/llm.",
        },
        {
          step: 2,
          title: "Add or Edit Provider",
          detail:
            "Click 'Add Provider' to open the configuration dialog. Choose your Provider Type (OpenAI, Anthropic, or Custom).",
          tip: "For local models via Ollama, select Custom and use http://127.0.0.1:11434/v1.",
        },
        {
          step: 3,
          title: "Input API Key and Model ID",
          detail: "Enter your API key and target model ID (e.g. gpt-4o, claude-3-7-sonnet-20250219, or qwen2.5:14b).",
        },
        {
          step: 4,
          title: "Test Connection & Save",
          detail: "Click 'Test Connection' to perform a live token handshake. Once confirmed green, click Save.",
        },
      ]}
      configFields={[
        {
          field: "Provider Type",
          type: "enum",
          description: "API schema protocol format (OpenAI, Anthropic, Custom Compatible).",
          defaultVal: "OpenAI",
        },
        {
          field: "Base URL",
          type: "string",
          description: "Upstream API endpoint URL prefix (can point to local proxy or gateway).",
          defaultVal: "https://api.openai.com/v1",
        },
        {
          field: "API Key",
          type: "secret",
          description: "Provider credential key, encrypted at rest using AES-GCM in PostgreSQL.",
        },
        {
          field: "Default Temperature",
          type: "float",
          description: "Sampling randomness applied to autonomous planning sessions.",
          defaultVal: "0.2",
        },
      ]}
      gotchas={[
        {
          title: "Context Window Exhaustion",
          description:
            "For long multi-step pen tests, choose models with at least 64k tokens context window (like GPT-4o or Claude 3.5 Sonnet) to prevent truncation of tool logs.",
        },
        {
          title: "Model Tool Calling Compatibility",
          description:
            "Ensure your selected model natively supports JSON function/tool calling. Models without tool call token support will fail autonomous loops.",
        },
      ]}
    />
  );
}
