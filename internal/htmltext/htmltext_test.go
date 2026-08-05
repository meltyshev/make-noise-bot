package htmltext

import "testing"

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
			name:     "inline formatting is preserved",
			fragment: "<b>жирный</b> и <i>курсив</i>",
			want:     "<b>жирный</b>и<i>курсив</i>",
		},
		{
			name:     "links are absolutized",
			fragment: `<a href="/foo">линк</a>`,
			want:     `<a href="https://example.com/foo">линк</a>`,
		},
		{
			name:     "images become urls",
			fragment: `до <img src="pic.png"> после`,
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
			name:     "spaces collapse and newlines squeeze",
			fragment: "<div></div><div></div><div>много    пробелов</div>",
			want:     "много пробелов",
		},
		{
			// Engine HTML sometimes carries JSON-escaped quotes into href
			// values.
			name:     "escaped quote artifact in urls is stripped",
			fragment: `<a href='/foo\"'>линк</a>`,
			want:     `<a href="https://example.com/foo">линк</a>`,
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
