"use client";

import { Sparkles } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function SkillDocPage() {
  return (
    <DocModuleTemplate
      title="Skill"
      description="Inspect specialized penetration testing skill packs, SKILL.md specs, and bundled tool configurations."
      groupLabel="SYSTEM"
      icon={Sparkles}
      consoleRoute="/system/skills"
      badge="Tool Packs"
      purpose="Organizes collections of individual tools, domain knowledge rules, and workflow guidance into reusable Skill Packs. Examples include 'api-recon' for OpenAPI mapping, 'scopesentry' for asset discovery, and 'playwright-cli' for browser crawling."
      prerequisites={[
        "Skill directories located in the host skills/ folder containing a valid SKILL.md file.",
        "Underlying CLI binaries required by the skill installed on the system PATH.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Open Skill Directory",
          detail: "Click 'Skill' in the SYSTEM section of the sidebar or browse to /system/skills.",
        },
        {
          step: 2,
          title: "Browse Available Skill Packs",
          detail:
            "Review cards for each registered skill showing author, version, tool count, and operational capabilities.",
          tip: "Click on any skill card to read its full SKILL.md instructions and guidelines.",
        },
        {
          step: 3,
          title: "Inspect Exported Tools",
          detail: "Verify which CLI commands are bundled inside the skill pack and check their parameter schemas.",
        },
        {
          step: 4,
          title: "Enable or Disable Skills",
          detail:
            "Toggle skill activation switches to control which skills are globally exposed to agent reasoning loops.",
        },
      ]}
      configFields={[
        {
          field: "Skill Name",
          type: "string",
          description: "Unique folder identifier under the skills/ directory matching frontmatter name.",
        },
        {
          field: "Enabled Status",
          type: "boolean",
          description: "Toggles whether this skill pack is active and selectable by agents.",
          defaultVal: "true",
        },
        {
          field: "Auto Reload on Save",
          type: "boolean",
          description: "Automatically parses updated SKILL.md files on the filesystem without restarting the daemon.",
          defaultVal: "true",
        },
        {
          field: "Safety Gating Profile",
          type: "enum",
          description: "Default execution risk level assigned to all tools bundled inside this skill pack.",
          defaultVal: "Standard",
        },
      ]}
      gotchas={[
        {
          title: "Missing SKILL.md Manifest",
          description:
            "If a newly added skill folder does not appear in the UI, check that it contains a valid SKILL.md markdown file with YAML frontmatter.",
        },
        {
          title: "Dependency Missing in PATH",
          description:
            "Even if a skill is enabled, tools will fail execution if their required CLI binaries are not installed on the host operating system.",
        },
      ]}
    />
  );
}
