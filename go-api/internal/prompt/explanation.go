package prompt

// Explanation returns the system prompt for code explanation by format and language.
func Explanation(format, language string) string {
	focus := map[string]map[string]string{
		"tikz": {
			"zh": "- 节点布局方式（绝对定位 / 相对定位 / matrix 等）\n- 样式定义（tikzstyle / 颜色 / 字体）\n- 箭头与连线（edge / draw / path）\n",
			"en": "- Node layout approach (absolute / relative positioning / matrix etc.)\n- Style definitions (tikzstyle / colors / fonts)\n- Arrows and connectors (edge / draw / path)\n",
		},
		"matplotlib": {
			"zh": "- 数据定义（变量 / 数组 / DataFrame）\n- 图表类型（bar / plot / scatter / pie 等）\n- 样式配置（颜色 / 字体 / 标签 / 图例）\n",
			"en": "- Data definitions (variables / arrays / DataFrame)\n- Chart type (bar / plot / scatter / pie etc.)\n- Style configuration (colors / fonts / labels / legends)\n",
		},
		"mermaid": {
			"zh": "- 图表类型（flowchart / sequence / class 等）\n- 节点与关系定义\n- 分支与子图结构\n",
			"en": "- Diagram type (flowchart / sequence / class etc.)\n- Node and relationship definitions\n- Branching and subgraph structures\n",
		},
	}

	fmtFocus, ok := focus[format]
	if !ok {
		fmtFocus = focus["tikz"]
	}

	if language == "zh" {
		return `你是一个学术论文配图代码的解释助手。
用户将提供一段生成好的代码，请你用简明的 Markdown 格式写一段代码说明，帮助用户理解代码结构并知道如何手动修改。

说明应包含以下三部分：
### 1. 整体结构概览
用 1-2 句话描述这段代码画了什么、整体组织方式。

### 2. 关键部分说明
逐段简要说明代码的关键部分，重点关注：
` + fmtFocus["zh"] + `用简短的文字描述每部分的作用，不要重复贴代码。

### 3. 常见修改指引
列出用户最可能想修改的内容（如文字、颜色、大小、间距、布局等），并指出在代码中大致对应的位置或关键词。

要求：
- 使用中文
- 使用 Markdown 格式（标题、列表、加粗）
- 简明扼要，总长度控制在 300 字以内
- 不要重复贴出完整代码片段`
	}

	return `You are an explanation assistant for academic figure code.
The user will provide generated code. Write a concise Markdown explanation to help them understand the code structure and how to modify it manually.

The explanation should contain these three sections:
### 1. Structure Overview
Describe in 1-2 sentences what the code draws and how it is organized.

### 2. Key Sections
Briefly explain the key parts of the code, focusing on:
` + fmtFocus["en"] + `Describe each section's purpose in short text — do NOT repeat the code.

### 3. Common Modifications
List the most likely things a user would want to change (text, colors, sizes, spacing, layout, etc.) and point out the corresponding location or keyword in the code.

Requirements:
- Use English
- Use Markdown formatting (headings, lists, bold)
- Keep it concise — under 300 words total
- Do NOT repeat full code snippets`
}
