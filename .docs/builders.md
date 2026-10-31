# Builders

The `gda/builder` package provides fluent constructors for Discord embeds, action rows, buttons, select menus, and modals. All builders validate character limits and component layout rules at `.Build()` time.

## Embeds

```go
embed, err := builder.NewEmbed().
	SetTitle("title").
	SetDescription("description").
	SetColor(0xFFFFFF).
	SetURL("https://sfossl.org").
	AddField("name", "value", false).
	SetAuthor("author", "https://sfossl.org", "").
	SetFooter("footer", "").
	SetTimestamp(time.Now()).
	Build()
if err != nil {
	log.Fatalf("invalid embed: %v", err)
}
```

Constraints validated at build time:
- Title: Max 256 characters (`ErrEmbedTitleTooLong`)
- Description: Max 4096 characters (`ErrEmbedDescriptionTooLong`)
- Fields: Max 25 fields (`ErrEmbedTooManyFields`)
- Field Name: Non-empty, max 256 characters (`ErrEmbedFieldNameTooLong`, `ErrEmbedFieldNameEmpty`)
- Field Value: Non-empty, max 1024 characters (`ErrEmbedFieldValueTooLong`, `ErrEmbedFieldValueEmpty`)
- Footer Text: Max 2048 characters (`ErrEmbedFooterTooLong`)
- Author Name: Max 256 characters (`ErrEmbedAuthorTooLong`)
- Total Combined Runes: Max 6000 characters (`ErrEmbedTotalLengthExceeded`)

## Buttons

```go
// interactive button
btn, err := builder.NewButton().
	SetStyle(builder.ButtonStylePrimary).
	SetCustomID("btn_action").
	SetLabel("submit").
	Build()

// external link button
linkBtn, err := builder.NewButton().
	SetStyle(builder.ButtonStyleLink).
	SetURL("https://sfossl.org").
	SetLabel("website").
	Build()

// action row container (up to 5 buttons per row)
row, err := builder.NewActionRow(btn, linkBtn)
```

Button styles:
- `ButtonStylePrimary` (Blurple)
- `ButtonStyleSecondary` (Grey)
- `ButtonStyleSuccess` (Green)
- `ButtonStyleDanger` (Red)
- `ButtonStyleLink` (Grey with external URL icon)

## Selects

```go
selectMenu, err := builder.NewStringSelect("select_item").
	SetPlaceholder("choose an item").
	SetValues(1, 1).
	AddOption("first", "val_1", "description for first", false).
	AddOption("second", "val_2", "description for second", false).
	Build()

row, err := builder.NewActionRow(selectMenu)
```

Constraints:
- Custom ID: Max 100 characters (`ErrSelectCustomIDTooLong`)
- Placeholder: Max 150 characters (`ErrSelectPlaceholderTooLong`)
- Options: Max 25 options (`ErrSelectTooManyOptions`)
- Action Row: Select menus must be the only component in an action row (`ErrActionRowSelectNotAlone`)

## Modals

Construct pop-up input dialogs:

```go
input := builder.NewTextInput("input_reason", "reason", builder.TextInputStyleParagraph).
	SetPlaceholder("enter details").
	SetLength(10, 200)

modal, err := builder.NewModal("modal_report", "report form").
	AddTextInput(input).
	Build()
```

Previous: [Rate Limits](rate-limits.md)  
Next: [Types & Models](types.md)
