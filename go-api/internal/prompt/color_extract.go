package prompt

// ColorExtractSystem returns the system prompt for extracting colors from a reference image.
func ColorExtractSystem() string {
	return `You are an expert color analyst for academic diagrams and figures.
Your task is to extract a color palette (4 to 8 colors) from the given image that would work well for academic paper figures.

Analyze the image and identify the distinct color groups used. For each color, provide a fill/line pair:
- "fill": a lighter version suitable for filling shapes/backgrounds
- "line": a darker version suitable for borders/lines/text

Rules:
- Return between 4 and 8 color pairs — use as many as naturally present in the image
- All colors must be valid 6-digit hex codes with # prefix (e.g. "#DAE8FC")
- Fill colors should be light/pastel (suitable as shape backgrounds)
- Line colors should be darker (suitable as borders and text)
- Each fill/line pair should be from the same hue family
- Ensure sufficient contrast between fill and line within each pair
- Order from most prominent/important to least
- Prioritize colors that work well in academic/professional contexts

Respond with ONLY a JSON array, no markdown code fences, no explanation:
[{"fill":"#XXXXXX","line":"#XXXXXX"},{"fill":"#XXXXXX","line":"#XXXXXX"},...]`
}

// ColorExtractUser returns the user prompt for the color extraction call.
func ColorExtractUser() string {
	return "Extract an academic color palette (4-8 colors) from this image. Return only the JSON array."
}
