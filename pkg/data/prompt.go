package data

import (
	"fmt"
	"strings"
)

type PromptBuilder struct {
	sb    strings.Builder
	dirty bool
}

func NewPrompt() *PromptBuilder {
	return &PromptBuilder{}
}

func (b *PromptBuilder) Header(level int, text string) *PromptBuilder {
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	b.gap()
	b.sb.WriteString(strings.Repeat("#", level))
	b.sb.WriteByte(' ')
	b.sb.WriteString(strings.TrimSpace(text))
	b.sb.WriteByte('\n')
	b.dirty = true
	return b
}

func (b *PromptBuilder) Section(text string) *PromptBuilder {
	t := strings.TrimSpace(text)
	if t == "" {
		return b
	}
	b.gap()
	b.sb.WriteString(t)
	b.sb.WriteByte('\n')
	b.dirty = true
	return b
}

func (b *PromptBuilder) Sectionf(format string, a ...any) *PromptBuilder {
	return b.Section(fmt.Sprintf(format, a...))
}

func (b *PromptBuilder) Bullet(text string) *PromptBuilder {
	b.sb.WriteString("- ")
	b.sb.WriteString(strings.TrimSpace(text))
	b.sb.WriteByte('\n')
	b.dirty = true
	return b
}

func (b *PromptBuilder) KV(key, value string) *PromptBuilder {
	b.sb.WriteString("- **")
	b.sb.WriteString(key)
	b.sb.WriteString("**: ")
	b.sb.WriteString(value)
	b.sb.WriteByte('\n')
	b.dirty = true
	return b
}

func (b *PromptBuilder) Code(lang, body string) *PromptBuilder {
	b.gap()
	b.sb.WriteString("```")
	b.sb.WriteString(lang)
	b.sb.WriteByte('\n')
	b.sb.WriteString(strings.TrimRight(body, "\n"))
	b.sb.WriteString("\n```\n")
	b.dirty = true
	return b
}

func (b *PromptBuilder) String() string {
	return strings.TrimSpace(b.sb.String())
}

func (b *PromptBuilder) gap() {
	if b.dirty {
		b.sb.WriteByte('\n')
	}
}
