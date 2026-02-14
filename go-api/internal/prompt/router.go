package prompt

import "fmt"

// Router returns the system prompt for the text-analysis router agent (step 1).
// It recommends 1-3 figure types WITHOUT generating drawing_prompt.
func Router(language, thesisTitle, thesisAbstract string) string {
	langLabel := "English"
	langConstraints := `- Use English.
- Do not use obscure or abstract jargon.
- Do not use exaggerated words.
- Must be fluent and logical.`
	if language == "zh" {
		langLabel = "Chinese"
		langConstraints = `- 使用简体中文
- 不要使用引号
- 不要使用括号补充英语内容
- 不要使用各种难懂抽象的词语
- 不要使用夸大词语
- 需要流畅，有逻辑`
	}

	thesisBlock := ""
	if thesisTitle != "" || thesisAbstract != "" {
		thesisBlock = "\n<thesis>\n"
		if thesisTitle != "" {
			thesisBlock += fmt.Sprintf("Title: %s\n", thesisTitle)
		}
		if thesisAbstract != "" {
			thesisBlock += fmt.Sprintf("Abstract: %s\n", thesisAbstract)
		}
		thesisBlock += "</thesis>\n"
	}

	return fmt.Sprintf(`<constraints>
%s
</constraints>

<task>
- First, read the content in <data> carefully. Identify the academic domain and adopt the identity of a senior expert in that field.
- If <thesis> is provided, also read the thesis title and abstract to understand the broader context.
- Recommend EXACTLY 3 different figures that could illustrate the content from different angles.
  Think about: architecture/structure diagrams, flow/process diagrams, comparison/relationship diagrams, data flow diagrams, etc.
  Each figure should have a distinct purpose — do NOT recommend variations of the same figure.
- Output a JSON array with exactly 3 elements (no markdown fences, no extra text). Each element:
  {"identity": "...", "title": "...", "description": "...", "priority": 1}

  - "identity": your adopted expert identity (e.g. "区块链安全研究员", "AI系统架构师")
  - "title": short figure title (e.g. "系统整体架构图", "训练流程图")
  - "description": 1-2 sentence %s description of what this figure should show and why it helps the reader
  - "priority": 1 = most recommended, 2 = second, 3 = third

Do NOT include a "drawing_prompt" field. That will be generated in a separate step.
You MUST output exactly 3 recommendations in the JSON array.
</task>
%s`, langConstraints, langLabel, thesisBlock)
}

// RouterDrawingPrompt returns the system prompt for generating a detailed drawing_prompt (step 2).
// Called after the user selects a recommended figure.
// colorDefs is the full \definecolor block with hex values, so the LLM can read actual colors.
func RouterDrawingPrompt(language, thesisTitle, thesisAbstract, colorDefs string) string {
	langLabel := "English"
	langConstraints := `- Use English.
- Do not use obscure or abstract jargon.
- Do not use exaggerated words.
- The drawing instructions must be extremely detailed.
- Must be fluent and logical.
- The figure should not be designed too wide.`
	if language == "zh" {
		langLabel = "Chinese"
		langConstraints = `- 使用简体中文
- 不要使用引号
- 不要使用括号补充英语内容
- 不要使用各种难懂抽象的词语
- 不要使用夸大词语
- 需要非常详细的画图指令
- 需要流畅，有逻辑`
	}

	thesisBlock := ""
	if thesisTitle != "" || thesisAbstract != "" {
		thesisBlock = "\n<thesis>\n"
		if thesisTitle != "" {
			thesisBlock += fmt.Sprintf("Title: %s\n", thesisTitle)
		}
		if thesisAbstract != "" {
			thesisBlock += fmt.Sprintf("Abstract: %s\n", thesisAbstract)
		}
		thesisBlock += "</thesis>\n"
	}

	return fmt.Sprintf(`<constraints>
%s
</constraints>

<palette>
%s
</palette>

<task>
You are given a figure recommendation (title + description) and the original content in <data>.
If <thesis> is provided, also read the thesis title and abstract to understand the broader context.
Adopt the expert identity specified in <figure>.

Your sole task: generate extremely detailed %s drawing instructions for the specified figure.

Output ONLY the drawing instructions as plain text (NOT JSON, NOT markdown). Write as flowing prose with multiple paragraphs.

CRITICAL REQUIREMENTS:
- At least 500 words, ideally 800+.
- Describe every visual element: nodes, boxes, arrows, labels, groupings, layers.
- Specify spatial relationships: what is above/below/left/right of what.
- Describe connection logic: which arrows connect which elements, direction, labels on arrows.
- Include all label text that should appear in the figure.
- Describe the overall layout structure: top-to-bottom flow, left-to-right pipeline, layered architecture, etc.
- The code-generation agent will ONLY see your text — nothing else. So nothing can be omitted.
- Do NOT compress or summarize. Be exhaustive and thorough.

COLOR ASSIGNMENT:
Read <palette> above — those are hex color definitions. Understand what actual colors they represent (blues, greens, oranges, purples, reds, greys, etc.).
For each visual element, state what color to use in plain language.
Example: "编码器模块使用蓝色填充，解码器模块使用绿色填充，注意力层使用橙色填充，输出层使用紫色填充，背景容器使用灰色。"
Use plain color words (蓝色, 绿色, 橙色, 紫色, 红色, 灰色 / blue, green, orange, purple, red, grey). Do NOT use technical names like primaryFill or secondaryLine.
</task>
%s`, langConstraints, colorDefs, langLabel, thesisBlock)
}
