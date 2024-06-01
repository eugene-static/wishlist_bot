package format

import "fmt"

const (
	Bold          = "bold"
	Italic        = "italic"
	Underline     = "underline"
	Strikethrough = "strikethrough"
	Spoiler       = "spoiler"
	Monotype      = "monotype"
	Code          = "code"
)

func Format(text string, style string) string {
	switch style {
	case Bold:
		text = fmt.Sprintf("<b>%s</b>", text)
	case Italic:
		text = fmt.Sprintf("<i>%s</i>", text)
	case Underline:
		text = fmt.Sprintf("<u>%s</u>", text)
	case Strikethrough:
		text = fmt.Sprintf("<s>%s</s>", text)
	case Spoiler:
		text = fmt.Sprintf("<span class=\"tg-spoiler\">%s</span>", text)
	case Monotype:
		text = fmt.Sprintf("<code>%s</code>", text)
		//case Code:
		//	text = fmt.Sprintf("```%s```", text)
	}
	return text
}
