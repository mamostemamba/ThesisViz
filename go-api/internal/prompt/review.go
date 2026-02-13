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
7. 连线穿越文字或节点（线条/箭头穿过其他方块或文字区域，这是严重问题）
8. 斜线交叉（非相邻节点之间使用了对角线直连，导致线条交叉混乱）

8.5. 原始代码泄露（节点内显示了 LaTeX 命令原文，如 \textbf{...}、\begin{tabular}、\footnotesize 等反斜杠命令，而非正常渲染的文字）——这是严重问题

二、内容完整性检查项（如果用户提供了画图提示词，必须对照检查）：
9. 内容缺失（提示词中要求的模块、节点、标签、文字等在图中缺失）
10. 内容不符（图中的元素与提示词描述的不一致，如名称错误、关系错误）
11. 文字内容缺失（节点或标签只有框没有文字，或文字为占位符）

请严格以如下 JSON 格式回复，不要输出任何其他内容：
{"passed": true/false, "score": 8, "issues": ["问题描述1", "问题描述2", ...], "critique": "1-3句整体点评"}

score 评分标准（1-10 分）：
- 9-10: 布局优秀，排版精美，可以直接用于论文发表
- 7-8: 基本合格，有小瑕疵但可以修复
- 4-6: 有明显问题但整体结构可辨认，需要修复
- 1-3: 布局混乱（节点严重重叠、线条大面积交叉、"意面图"），必须推倒重画

passed 判定规则：
- score >= 7 时 passed=true
- score < 7 时 passed=false

- 如果 passed=true，critique 写积极评价，例如"图表布局清晰，内容完整，整体质量良好。"
- 如果 passed=false，critique 写整体概述+问题方向，例如"图表整体结构合理，但存在以下问题需要修复：……"
- 如果图片质量合格且内容完整，返回 {"passed": true, "score": 9, "issues": [], "critique": "..."}。

三、细线条与装饰元素专项检查：
12. 特别注意大括号（curly braces）、虚线、细线条装饰等元素——它们在图片压缩后容易变模糊甚至消失
13. 如果提示词中要求了分组括号但图中看不到清晰的粗黑括号线条，报告"装饰元素过细或缺失"
14. 不要因为代码中写了 \draw[decorate] 就假设图上一定显示了——必须在图片中实际确认

审查严格度要求：
- 原始代码泄露（节点中可见反斜杠命令）= 必须报告（严重问题）
- 连线穿越节点/文字 = 必须报告（严重问题）
- 斜线交叉混乱 = 必须报告（严重问题）
- 文字重叠或截断 = 必须报告
- 内容缺失 = 必须报告（最高优先级）
- 装饰元素缺失或不可见 = 必须报告
- 只有在没有以上严重问题时才可以 passed=true
- 轻微的间距不均匀、对齐偏差可以忽略
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
7. Lines crossing over nodes/text (arrows or lines pass through other boxes or text regions — this is a SEVERE issue)
8. Diagonal line crossings (non-adjacent nodes connected with diagonal straight lines, causing visual clutter)

8.5. Raw code leakage (nodes display raw LaTeX commands such as \textbf{...}, \begin{tabular}, \footnotesize instead of properly rendered text) — this is a SEVERE issue

B. Content completeness checks (if a drawing prompt is provided, compare against it):
9. Missing content (modules, nodes, labels, or text required by the prompt are absent)
10. Incorrect content (elements don't match the prompt — wrong names, wrong relationships)
11. Missing text (nodes or labels are empty boxes or have placeholder text)

Respond strictly in this JSON format with no other text:
{"passed": true/false, "score": 8, "issues": ["issue description 1", "issue description 2", ...], "critique": "1-3 sentence overall assessment"}

Score rubric (1-10):
- 9-10: Excellent layout, publication-ready quality
- 7-8: Acceptable with minor fixable issues
- 4-6: Noticeable problems but overall structure is recognizable, needs fixing
- 1-3: Chaotic layout (severe node overlaps, lines crossing everywhere, "spaghetti diagram"), must be completely redrawn

Passed rule:
- score >= 7 → passed=true
- score < 7 → passed=false

- If passed=true, write a positive critique, e.g. "The figure layout is clean, content is complete, and overall quality is good."
- If passed=false, write an overview with problem direction, e.g. "The figure structure is reasonable, but the following issues need to be fixed: ..."
- If the image quality is acceptable and content is complete, return {"passed": true, "score": 9, "issues": [], "critique": "..."}.

C. Thin lines and decorative elements (special attention):
12. PAY SPECIAL ATTENTION to thin lines and decorative elements like CURLY BRACES, dotted lines, and decoration lines — they are often lost in image compression
13. If the prompt requires grouping brackets but you don't clearly see a thick black bracket in the image, report "Visual element too thin or missing"
14. Do NOT assume decorative elements exist just because the code contains \draw[decorate] — you MUST visually confirm in the image

Strictness requirements:
- Raw code leakage (backslash commands visible in nodes) = MUST report (severe)
- Lines crossing over nodes/text = MUST report (severe)
- Diagonal line crossings/clutter = MUST report (severe)
- Text overlap or truncation = MUST report
- Missing content = MUST report (highest priority)
- Missing or invisible decorative elements = MUST report
- Only set passed=true when NONE of the above severe issues exist
- Minor spacing unevenness or slight alignment offsets can be ignored
Content-missing issues have HIGHER priority than layout issues — report them first.`
}

// ReviewFix returns the user prompt to fix issues found during visual review.
func ReviewFix(issues []string, score float64, language, drawingPrompt string) string {
	issueText := strings.Join(issues, "\n- ")
	issueText = "- " + issueText

	fixTechniques := `
TikZ 修复技巧（请优先使用）：
- 节点重叠/错位 → 如果还没用 \matrix，请重构为 \matrix (m) [matrix of nodes, row sep=1.5cm, column sep=2cm, nodes={matrix_node}] { ... }; 这样网格化布局永远不会重叠。如果仍然拥挤，增大 row sep 到 2cm、column sep 到 2.5cm
- 线条穿越节点 → 改用曼哈顿路径: \draw[nice_arrow] (m-1-1) -| (m-3-3); 或 |- 语法
- 斜线交叉 → 拆成折线: (m-1-1.east) -- ++(0.5,0) |- (m-3-3.north);
- 文字重叠 → 增加 row sep / column sep，或调整 text width
- 布局溢出 → 缩小 minimum width/height，减小 column sep
- 元素遮挡 → 在 background layer 中绘制容器框: \begin{pgfonlayer}{background} ... \end{pgfonlayer}`

	fixTechniquesEN := `
TikZ fix techniques (use these as appropriate):
- Node overlap/misalignment → If not using \matrix yet, REFACTOR to: \matrix (m) [matrix of nodes, row sep=1.5cm, column sep=2cm, nodes={matrix_node}] { ... }; Grid layout guarantees no overlap. If still crowded, increase row sep to 2cm and column sep to 2.5cm.
- Lines crossing nodes → Use Manhattan routing: \draw[nice_arrow] (m-1-1) -| (m-3-3); or |- syntax
- Diagonal crossings → Break into segments: (m-1-1.east) -- ++(0.5,0) |- (m-3-3.north);
- Text overlap → Increase row sep / column sep, or adjust text width
- Layout overflow → Reduce minimum width/height, decrease column sep
- Element occlusion → Draw containers on background layer: \begin{pgfonlayer}{background} ... \end{pgfonlayer}`

	twoStepFormat := `
请按以下格式输出：

=== REASONING ===
对每个问题，分析：
1. 上一版为什么出现了这个问题（代码中哪行/哪个参数导致的）
2. 具体的修复方案（改什么参数，改成多少）

=== FIXED CODE ===
完整的修复后 TikZ 代码（从 \begin{tikzpicture} 到 \end{tikzpicture}）

修复要求：
- 针对每个问题至少修改一个具体参数（间距、尺寸、位置等）
- 代码必须与原版有明显差异，禁止只改注释或空白
- 修复排版时不要删除或遗漏任何文字内容、节点或模块`

	twoStepFormatEN := `
Please use the following output format:

=== REASONING ===
For each issue, analyze:
1. Why did this problem occur in the previous version (which line/parameter caused it)
2. The specific fix (which parameter to change, to what value)

=== FIXED CODE ===
The complete fixed TikZ code (from \begin{tikzpicture} to \end{tikzpicture})

Fix requirements:
- For each issue, change at least one concrete parameter (spacing, size, position, etc.)
- The code MUST differ visibly from the original — do NOT only change comments or whitespace
- Do NOT remove or omit any text content, nodes, or modules while fixing layout`

	if language == "zh" {
		p := fmt.Sprintf("经过视觉审查，当前分数: %.0f/10，发现以下问题：\n\n%s\n\n", score, issueText)
		if score < 6 {
			p += "分数较低，请重点关注整体布局结构，确保节点不重叠、连线不穿越文字。\n\n"
		}
		if drawingPrompt != "" {
			p += fmt.Sprintf("原始画图要求如下，修复时必须确保所有内容完整保留：\n%s\n\n", drawingPrompt)
		}
		p += "请修复这些问题。" + fixTechniques + twoStepFormat
		return p
	}

	p := fmt.Sprintf("A visual review scored this %.0f/10. Issues found:\n\n%s\n\n", score, issueText)
	if score < 6 {
		p += "Score is low — focus on overall layout structure: ensure no node overlaps and no lines crossing through text.\n\n"
	}
	if drawingPrompt != "" {
		p += fmt.Sprintf("Original drawing requirements (all content must be preserved):\n%s\n\n", drawingPrompt)
	}
	p += "Please fix these issues." + fixTechniquesEN + twoStepFormatEN
	return p
}
