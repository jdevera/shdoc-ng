# HTML Template Theming Guide

The built-in HTML template (`html.tmpl`) uses a three-layer CSS custom-property
architecture designed so that creating a custom theme requires touching as few
lines as possible.

---

## The Three Layers

```
Layer 1 — Palette       --ctp-*      Raw hex values per colour scheme
    ↓
Layer 2 — Semantic      --theme-*    Role-based aliases mapped from the palette
    ↓
Layer 3 — Components    CSS rules    Use only --theme-* — never palette tokens
```

### Layer 1 — Palette (`--ctp-*`)

Each Catppuccin flavour defines its raw colour tokens under a
`[data-theme="<name>"]` selector. Hex values appear **only** in this layer.

```css
[data-theme="mocha"] {
  --ctp-base:    #1e1e2e;
  --ctp-text:    #cdd6f4;
  --ctp-sky:     #89dceb;
  /* … */
}
```

The full palette for each flavour contains these tokens:

| Token             | Role in Catppuccin |
|-------------------|--------------------|
| `--ctp-base`      | Page background |
| `--ctp-mantle`    | Slightly darker background (sidebar) |
| `--ctp-crust`     | Darkest background (code blocks) |
| `--ctp-surface0`  | Elevated surface (inputs, inline code bg) |
| `--ctp-surface1`  | Higher surface (borders, separators) |
| `--ctp-surface2`  | Highest surface |
| `--ctp-overlay0`  | Subdued text (labels, placeholders) |
| `--ctp-overlay1`  | Muted text |
| `--ctp-overlay2`  | Less muted text |
| `--ctp-subtext0`  | Secondary text |
| `--ctp-subtext1`  | Near-primary secondary text |
| `--ctp-text`      | Primary body text |
| `--ctp-rosewater` | Accent |
| `--ctp-flamingo`  | Accent |
| `--ctp-pink`      | Accent |
| `--ctp-mauve`     | Purple accent (used for `@section`) |
| `--ctp-red`       | Error / deprecated |
| `--ctp-maroon`    | Softer red |
| `--ctp-peach`     | Orange accent (types) |
| `--ctp-yellow`    | Warning |
| `--ctp-green`     | Success / options |
| `--ctp-teal`      | Teal accent |
| `--ctp-sky`       | Sky blue (function names, inline code) |
| `--ctp-sapphire`  | Medium blue |
| `--ctp-blue`      | Blue (links) |
| `--ctp-lavender`  | Light purple (titles, active TOC) |

---

### Layer 2 — Semantic (`--theme-*`)

The semantic layer is defined **once** for all themes in `[data-theme]` (no
flavour suffix). It maps a *purpose* to a palette token. Because this block is
shared, every flavour automatically adopts the same mapping.

```css
[data-theme] {
  --theme-bg:          var(--ctp-base);
  --theme-fn-name:     var(--ctp-sky);
  /* … */
}
```

Full semantic variable reference:

| Variable                    | Used for |
|-----------------------------|----------|
| `--theme-bg`                | Page background |
| `--theme-bg-sidebar`        | Sidebar / panel background |
| `--theme-bg-code`           | Fenced code block background |
| `--theme-surface`           | Inputs, inline code background, button background |
| `--theme-border`            | Default dividers and section borders |
| `--theme-border-faint`      | Subtle separators, scrollbar tracks |
| `--theme-text`              | Primary body copy |
| `--theme-text-secondary`    | Descriptions, table cell content |
| `--theme-text-muted`        | Section labels, metadata |
| `--theme-text-faint`        | Placeholders, anchor glyphs, list bullets |
| `--theme-accent`            | File title, active TOC item, focus rings |
| `--theme-fn-name`           | Function name headings |
| `--theme-section`           | `@section` headings |
| `--theme-link`              | Hyperlinks |
| `--theme-code-text`         | Inline code foreground |
| `--theme-type`              | Type badges (`string`, `int`, …) |
| `--theme-option`            | Option flag tokens |
| `--theme-deprecated`        | Deprecated badge border and label |
| `--theme-deprecated-subtle` | Deprecated reason text |
| `--theme-warning`           | `@warning` list items |
| `--theme-success`           | Copy-button success flash |

---

### Layer 3 — Components

All CSS component rules consume only `--theme-*` variables. No palette token
(`--ctp-*`) is referenced here. This guarantees that any change in Layer 1 or
Layer 2 is enough to retheme the entire page.

---

## Creating a Custom Theme

There are two approaches depending on how much you want to reuse.

### Approach A — New palette, reuse semantic mapping

Define `--ctp-*` tokens for your theme name. The semantic-mapping block in
`[data-theme]` picks them up automatically.

```css
[data-theme="nord"] {
  --ctp-base:     #2e3440;
  --ctp-mantle:   #272c36;
  --ctp-crust:    #1e2329;
  --ctp-surface0: #3b4252;
  --ctp-surface1: #434c5e;
  --ctp-surface2: #4c566a;
  --ctp-overlay0: #616e88;
  --ctp-overlay1: #677691;
  --ctp-overlay2: #7b88a1;
  --ctp-subtext0: #8896af;
  --ctp-subtext1: #9aa5be;
  --ctp-text:     #eceff4;
  --ctp-mauve:    #b48ead;   /* @section headings */
  --ctp-red:      #bf616a;   /* deprecated */
  --ctp-maroon:   #d08770;
  --ctp-peach:    #d08770;   /* types */
  --ctp-yellow:   #ebcb8b;   /* warnings */
  --ctp-green:    #a3be8c;   /* options / success */
  --ctp-sky:      #88c0d0;   /* function names, inline code */
  --ctp-blue:     #81a1c1;   /* links */
  --ctp-lavender: #5e81ac;   /* accent */
  /* unused Catppuccin tokens — set to fallback values */
  --ctp-rosewater: #bf616a; --ctp-flamingo: #bf616a; --ctp-pink: #b48ead;
  --ctp-teal: #88c0d0; --ctp-sapphire: #81a1c1;
}
```

Then add the option to the `<select>` in `html.tmpl`:

```html
<option value="nord">Nord</option>
```

And initialise the default / add it to the allowed-flavours list in the script:

```js
const flavours = ['mocha', 'macchiato', 'frappe', 'latte', 'nord'];
```

---

### Approach B — Override semantic variables directly

Skip the palette entirely. Override `--theme-*` under your own selector.
You only need to list the variables whose values differ from another theme.

```css
/* Minimal monochrome override on top of Mocha */
[data-theme="mono"] {
  --ctp-base: #1a1a1a; --ctp-mantle: #111; --ctp-crust: #0a0a0a;
  --ctp-surface0: #2a2a2a; --ctp-surface1: #333; --ctp-surface2: #3a3a3a;
  --ctp-overlay0: #555; --ctp-overlay1: #666; --ctp-overlay2: #777;
  --ctp-subtext0: #888; --ctp-subtext1: #999; --ctp-text: #e0e0e0;
  --ctp-mauve: #aaa; --ctp-red: #cc5555; --ctp-maroon: #bb7777;
  --ctp-peach: #cc8855; --ctp-yellow: #bbaa55; --ctp-green: #77aa77;
  --ctp-sky: #7799bb; --ctp-blue: #6688cc; --ctp-lavender: #9999cc;
  --ctp-rosewater: #cc8888; --ctp-flamingo: #cc8888; --ctp-pink: #aa88aa;
  --ctp-teal: #6699aa; --ctp-sapphire: #7799bb;
}

/* Fine-tune individual semantic roles without touching the palette */
[data-theme="mono"] {
  --theme-accent:  #c0c0c0;
  --theme-fn-name: #b0b0b0;
  --theme-section: #a0a0a0;
}
```

---

## Quick-start checklist

1. Add a `[data-theme="<name>"]` block with your palette **or** semantic overrides.
2. Add `<option value="<name>">Display Name</option>` to the `<select>` in the template.
3. Add the theme name to the `flavours` array in the `<script>` block.
4. Test with `shdoc-ng --template html.tmpl -i yourscript.sh -o out.html`.

---

## Design rules to keep when customising the template itself

- **Layer 3 must not reference `--ctp-*` directly.** All component rules go
  through `--theme-*` only. If you need a new role, add a variable to
  Layer 2 first.
- **`color-mix()` for tinted surfaces is fine** — e.g.
  `color-mix(in srgb, var(--theme-accent) 8%, transparent)` — because it
  uses a semantic variable, not a palette one.
- **Hex values belong only in Layer 1.** This is what makes the system
  portable: a theme author only needs to read Layer 1 and Layer 2.
