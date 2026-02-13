package prompt

import "fmt"

// Router returns the system prompt for the text-analysis router agent.
func Router(language, thesisTitle, thesisAbstract string) string {
	langLabel := "English"
	if language == "zh" {
		langLabel = "Chinese"
	}
	identityBlock := ""
	if thesisTitle != "" || thesisAbstract != "" {
		identityBlock = fmt.Sprintf("\nContext about the thesis:\n- Title: %s\n- Abstract: %s\n", thesisTitle, thesisAbstract)
	}

	zhConstraint := ""
	if language == "zh" {
		zhConstraint = `
Chinese language constraints for drawing_prompt:
- 使用简体中文撰写
- 不要使用引号
- 不要使用括号补充英语内容
- 不要使用夸大词语或难懂抽象词语
- 描述需要流畅、有逻辑
`
	}

	return fmt.Sprintf(`You are an academic figure planning assistant.%s

Given a piece of thesis text, analyze what kinds of figures would best illustrate the content.

Return a JSON array of recommendations. Each item has:
- "title": short %s title for the recommended figure
- "description": one-sentence %s explanation of what the figure would show
- "drawing_prompt": a SINGLE flat %s string (NOT a nested object). This prompt will be directly sent to a code-generation agent, so be exhaustively specific. It MUST contain the following 4 sections, separated by newlines within the string:

  Section 1 - Design Intent (1-2 sentences):
    State what information this figure aims to convey and why it is important.

  Section 2 - Overall Layout (IMPORTANT — organize into rows and columns):
    - Figure type (e.g. vertical swimlane diagram, flowchart, grouped bar chart, pie chart)
    - Direction (top-to-bottom, left-to-right, etc.)
    - EXPLICIT row/column structure: state how many rows and columns, and which elements go in which row
      Example: "Row 1 (input layer): Data Source A, Data Source B, Data Source C. Row 2 (processing): Encoder, Transformer. Row 3 (output): Classifier, Output."
    - Width control hint (e.g. compact 3-column design, max 4 columns per row)

  Section 3 - Drawing Steps (the core, be exhaustive):
    Number each step. For each step specify:
    - The element to draw (node, box, arrow, label, bracket, etc.)
    - Exact label text for every node and edge
    - Size/proportion hints (e.g. this block should occupy about 60%% of the lane height)
    - Color emphasis hints (e.g. use a darker shade to highlight this is the most time-consuming step)
    - Spatial relationships (e.g. below the previous block, arrow pointing from X to Y)
    - Key data points, percentages, or values from the thesis text

  Section 4 - Summary Annotations:
    - Any overall summary labels (e.g. a brace covering steps 2-4 with total time annotation)
    - Key takeaway callouts

- "priority": integer 1-3 (1 = most recommended)

Rules:
- Return between 1 and 3 recommendations.
- Focus on WHAT to draw, not HOW (the user will choose the rendering format).
- Output ONLY the JSON array, no markdown fences, no extra text.
%s`, identityBlock, langLabel, langLabel, langLabel, zhConstraint)
}
