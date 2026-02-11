package prompt

import (
	"fmt"
	"strings"
)

// ReviewSystem returns the system prompt for visual quality review.
func ReviewSystem(language string) string {
	if language == "zh" {
		return `你是一个学术论文配图的质量审查专家。
你的任务是检查渲染后的图片是否存在排版问题以及内容是否完整。

一、排版检查项：
1. 文字重叠或截断（文字之间互相遮挡，或文字超出边框被裁掉）
2. 布局溢出（内容超出图片边界）
3. 间距问题（元素之间过于拥挤或间距不均匀）
4. 箭头或连线错位（箭头没有正确连接到目标元素）
5. 可读性问题（字体太小、颜色对比度不足）
6. 对齐问题（元素没有合理对齐，布局歪斜）

二、内容完整性检查项（如果用户提供了画图提示词，必须对照检查）：
7. 内容缺失（提示词中要求的模块、节点、标签、文字等在图中缺失）
8. 内容不符（图中的元素与提示词描述的不一致，如名称错误、关系错误）
9. 文字内容缺失（节点或标签只有框没有文字，或文字为占位符）

请严格以如下 JSON 格式回复，不要输出任何其他内容：
{"passed": true/false, "issues": ["问题描述1", "问题描述2", ...]}

如果图片质量合格且内容完整，返回 {"passed": true, "issues": []}。
只报告确实存在的明显问题，不要吹毛求疵。
内容缺失问题优先级高于排版问题，必须优先报告。`
	}

	return `You are a quality reviewer for academic paper figures.
Your task is to check the rendered image for layout issues AND content completeness.

A. Layout checks:
1. Text overlap or truncation (text obscured or clipped by boundaries)
2. Layout overflow (content extends beyond image boundaries)
3. Spacing issues (elements too crowded or unevenly spaced)
4. Arrow/connector misalignment (arrows not connecting to targets)
5. Readability issues (font too small, poor color contrast)
6. Alignment issues (elements not properly aligned, skewed layout)

B. Content completeness checks (if a drawing prompt is provided, compare against it):
7. Missing content (modules, nodes, labels, or text required by the prompt are absent)
8. Incorrect content (elements don't match the prompt — wrong names, wrong relationships)
9. Missing text (nodes or labels are empty boxes or have placeholder text)

Respond strictly in this JSON format with no other text:
{"passed": true/false, "issues": ["issue description 1", "issue description 2", ...]}

If the image quality is acceptable and content is complete, return {"passed": true, "issues": []}.
Only report clearly visible, obvious issues. Do not be overly picky.
Content-missing issues have HIGHER priority than layout issues — report them first.`
}

// ReviewFix returns the user prompt to fix issues found during visual review.
func ReviewFix(issues []string, language, drawingPrompt string) string {
	issueText := strings.Join(issues, "\n- ")
	issueText = "- " + issueText

	if language == "zh" {
		p := fmt.Sprintf("经过视觉审查，发现以下问题：\n\n%s\n\n", issueText)
		if drawingPrompt != "" {
			p += fmt.Sprintf("原始画图要求如下，修复时必须确保所有内容完整保留：\n%s\n\n", drawingPrompt)
		}
		p += "请修复这些问题。\n重要：修复排版时不要删除或遗漏任何文字内容、节点或模块。\n只输出完整的修复后代码，不要解释。"
		return p
	}

	p := fmt.Sprintf("A visual review found the following issues:\n\n%s\n\n", issueText)
	if drawingPrompt != "" {
		p += fmt.Sprintf("Original drawing requirements (all content must be preserved):\n%s\n\n", drawingPrompt)
	}
	p += "Please fix these issues.\nIMPORTANT: Do NOT remove or omit any text content, nodes, or modules while fixing layout.\nOutput ONLY the complete fixed code, no explanations."
	return p
}
