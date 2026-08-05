package htmltext

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{
			name:     "block tags become newlines",
			fragment: "<div>раз</div><p>два</p>три<br>четыре",
			want:     "раз\n\nдва\nтри\nчетыре",
		},
		{
			name:     "spaces around inline tags survive",
			fragment: "текст <b>жирный</b> и <i>курсив</i> хвост",
			want:     "текст <b>жирный</b> и <i>курсив</i> хвост",
		},
		{
			name:     "no space in source means none in output",
			fragment: "текст<b>жирный</b>",
			want:     "текст<b>жирный</b>",
		},
		{
			name:     "links are absolutized and keep spacing",
			fragment: `до <a href="/foo">линк</a> после`,
			want:     `до <a href="https://example.com/foo">линк</a> после`,
		},
		{
			name:     "images become urls",
			fragment: `до<img src="pic.png">после`,
			want:     "до https://example.com/base/pic.png после",
		},
		{
			name:     "self-closing img and br",
			fragment: `а<br/><img src="/x.jpg"/>`,
			want:     "а\nhttps://example.com/x.jpg",
		},
		{
			name:     "script and style are dropped",
			fragment: "до<script>alert(1)</script><style>b{}</style>после",
			want:     "допосле",
		},
		{
			name:     "text is escaped for telegram html",
			fragment: "<div>1 < 2 & 3 > 2</div>",
			want:     "1 &lt; 2 &amp; 3 &gt; 2",
		},
		{
			name:     "source newlines collapse into spaces",
			fragment: "первая\nвторая   строка",
			want:     "первая вторая строка",
		},
		{
			name:     "escaped quote artifact in urls is stripped",
			fragment: `<a href='/foo\"'>линк</a>`,
			want:     `<a href="https://example.com/foo">линк</a>`,
		},
		{
			name:     "unclosed tag is closed at the end",
			fragment: "<b>жирный до конца",
			want:     "<b>жирный до конца</b>",
		},
		{
			name:     "stray closing tag is dropped",
			fragment: "текст</b> хвост</i>",
			want:     "текст хвост",
		},
		{
			name:     "misnested tags are rebalanced",
			fragment: "<b>раз<i>два</b>три</i>",
			want:     "<b>раз<i>два</i></b><i>три</i>",
		},
		{
			name:     "nested link closes the previous one",
			fragment: `<a href="/1">один<a href="/2">два</a>`,
			want:     `<a href="https://example.com/1">один</a><a href="https://example.com/2">два</a>`,
		},
		{
			name:     "link without href leaves plain text",
			fragment: "до <a>линк</a> после",
			want:     "до линк после",
		},
		{
			name:     "extra tags map to the telegram set",
			fragment: "<strong>ж</strong><em>к</em><u>у</u><strike>с</strike><del>д</del><ins>и</ins><code>x</code>",
			want:     "<b>ж</b><i>к</i><u>у</u><s>с</s><s>д</s><u>и</u><code>x</code>",
		},
		{
			name:     "unknown tags disappear but text stays",
			fragment: `<font color="red">красный</font> <span>текст</span>`,
			want:     "красный текст",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.fragment, "https://example.com/base/"); got != tt.want {
				t.Errorf("Convert(%q) = %q, want %q", tt.fragment, got, tt.want)
			}
		})
	}
}

func TestStripTags(t *testing.T) {
	in := `текст <b>жирный</b> <a href="https://x">линк</a> 1 &lt; 2`
	if got, want := StripTags(in), "текст жирный линк 1 < 2"; got != want {
		t.Errorf("StripTags = %q, want %q", got, want)
	}
}

func balanced(t *testing.T, part string) {
	t.Helper()
	var stack []string
	for _, tok := range tokenize(part) {
		if !tok.isTag {
			continue
		}
		if tok.closing {
			if len(stack) == 0 || stack[len(stack)-1] != tok.name {
				t.Fatalf("unbalanced part: %q", part)
			}
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, tok.name)
		}
	}
	if len(stack) != 0 {
		t.Fatalf("unclosed tags in part: %q", part)
	}
}

func TestSplitShort(t *testing.T) {
	parts := Split("короткий <b>текст</b>", 100)
	if len(parts) != 1 || parts[0] != "короткий <b>текст</b>" {
		t.Errorf("parts = %q", parts)
	}
}

func TestSplitKeepsTagsBalanced(t *testing.T) {
	text := "<b>" + strings.Repeat("слово ", 100) + "</b>"
	parts := Split(text, 200)

	if len(parts) < 2 {
		t.Fatalf("expected several parts, got %d", len(parts))
	}
	totalWords := 0
	for _, part := range parts {
		if utf16Len(part) > 200 {
			t.Errorf("part exceeds limit: %d units", utf16Len(part))
		}
		balanced(t, part)
		if !strings.HasPrefix(part, "<b>") || !strings.HasSuffix(part, "</b>") {
			t.Errorf("bold not carried across parts: %q", part)
		}
		totalWords += strings.Count(part, "слово")
	}
	if totalWords != 100 {
		t.Errorf("words lost in split: %d of 100", totalWords)
	}
}

func TestSplitPrefersNewlines(t *testing.T) {
	text := strings.Repeat("строка номер такой-то\n", 30)
	for _, part := range Split(strings.TrimSpace(text), 100) {
		for _, line := range strings.Split(part, "\n") {
			if line != "строка номер такой-то" {
				t.Errorf("cut mid-line: %q", line)
			}
		}
	}
}

func TestSplitPreBlocks(t *testing.T) {
	text := "<pre>" + strings.Repeat("код\n", 50) + "</pre>"
	parts := Split(text, 80)
	if len(parts) < 2 {
		t.Fatalf("expected several parts, got %d", len(parts))
	}
	for _, part := range parts {
		if !strings.HasPrefix(part, "<pre>") || !strings.HasSuffix(part, "</pre>") {
			t.Errorf("pre not carried: %q", part)
		}
	}
}

func TestSplitCountsUTF16(t *testing.T) {
	parts := Split(strings.Repeat("\U0001F600", 10), 10)
	if len(parts) != 2 {
		t.Fatalf("emoji parts = %d, want 2", len(parts))
	}
	for _, part := range parts {
		if utf16Len(part) > 10 {
			t.Errorf("part exceeds utf16 limit: %q", part)
		}
	}
}

func TestSplitDoesNotCutEntities(t *testing.T) {
	text := strings.Repeat("x", 95) + "&amp;xyz"
	for _, part := range Split(text, 100) {
		if strings.Contains(part, "&am") && !strings.Contains(part, "&amp;") {
			t.Errorf("entity cut in half: %q", part)
		}
	}
}
