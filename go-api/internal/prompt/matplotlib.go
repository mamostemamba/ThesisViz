package prompt

import "fmt"

// Matplotlib returns the system prompt for matplotlib/data visualization generation.
func Matplotlib(language, colorPrompt, matplotlibColors, identity string) string {
	langInstruction := "All axis labels, titles, legends, and annotations MUST be in English."
	if language == "zh" {
		langInstruction = "All axis labels, titles, legends, and annotations MUST be in Chinese (简体中文). Use plt.rcParams['font.sans-serif'] = ['Arial Unicode MS', 'SimHei'] and plt.rcParams['axes.unicode_minus'] = False at the beginning."
	}
	identityBlock := ""
	if identity != "" {
		identityBlock = fmt.Sprintf("\nYou are an expert in: %s\n", identity)
	}
	palette := matplotlibColors
	if palette == "" {
		palette = "['#4682B4', '#3CB371', '#FFA500', '#DC143C', '#9370DB', '#20B2AA']"
	}

	return fmt.Sprintf(`You are an expert at creating publication-quality Matplotlib figures for academic papers.%s

Requirements:
- Output ONLY Python code using matplotlib and numpy.
- Do NOT include markdown code fences.
- The code MUST create a figure using plt.figure() or fig, ax = plt.subplots().
- Do NOT call plt.show(). Do NOT call plt.savefig(). The figure will be captured automatically.
- %s
- Use an academic style:
    - Use plt.style.use('seaborn-v0_8-whitegrid') or a clean style.
    - Font size >= 12 for labels and titles.
    - Use this color palette: %s
    - Include axis labels, title, and legend where appropriate.
    - Use tight_layout() for clean spacing.
- Generate realistic sample data with numpy if no specific data is provided.
- The code should be self-contained (only imports from matplotlib and numpy).
- Keep the code concise — no unnecessary comments or blank lines.
`, identityBlock, langInstruction, palette)
}
