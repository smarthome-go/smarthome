// Modified version of https://github.com/codemirror/theme-one-dark/blob/b2783a648d8d94e544993c7c1467dfea6ec86618/src/one-dark.ts

/// The editor theme styles for One Dark.
import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { EditorView } from '@codemirror/view'
import { tags as t } from '@lezer/highlight'
import "@fontsource/jetbrains-mono"

const highlightBackground = '#2c313a',
    tooltipBackground = '#353a42',
    //////////// Below colors taken from https://github.com/navarasu/onedark.nvim/blob/fdfe7bfff486acd102aae7fb2ff52e7e5f6c2bad/lua/onedark/palette.lua
    bg0 = '#282c34',
    bg3 = '#3b3f4c',
    bg_d = '#21252b',
    fg = '#abb2bf',
    purple = '#c678dd',
    green = '#98c379',
    orange = '#d19a66',
    blue = '#61afef',
    cyan = '#56b6c2',
    red = '#e86671',
    grey = '#5c6370',
    yellow = '#e1bd79',
    light_grey = 'rgba(255, 255, 255, 0.6)'

export const oneDarkTheme = EditorView.theme({
    '*': {
        fontFamily: 'Jetbrains Mono NL, monospace',
    },
    '&': {
        color: fg,
        backgroundColor: bg0,
        height: '100%',
    },
    '.cm-scroller': { overflow: 'auto' },

    '.cm-content': {
        caretColor: blue,
    },

    '.cm-cursor, .cm-dropCursor': { borderLeftColor: blue },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
        backgroundColor: bg3,
    },

    '.cm-panels': { backgroundColor: bg_d, color: fg },
    '.cm-panels.cm-panels-top': { borderBottom: '2px solid black' },
    '.cm-panels.cm-panels-bottom': { borderTop: '2px solid black' },

    '.cm-searchMatch': {
        backgroundColor: '#72a1ff59',
        outline: '1px solid #457dff',
    },
    '.cm-searchMatch.cm-searchMatch-selected': {
        backgroundColor: '#6199ff2f',
    },

    '.cm-activeLine': { backgroundColor: highlightBackground },
    '.cm-selectionMatch': { backgroundColor: '#aafe661a' },

    '&.cm-focused .cm-matchingBracket, &.cm-focused .cm-nonmatchingBracket': {
        backgroundColor: '#bad0f847',
        outline: '1px solid #515a6b',
    },

    '.cm-gutters': {
        backgroundColor: bg0,
        color: grey,
        border: 'none',
    },

    '.cm-activeLineGutter': {
        backgroundColor: highlightBackground,
    },

    '.cm-foldPlaceholder': {
        backgroundColor: 'transparent',
        border: 'none',
        color: '#ddd',
    },

    '.cm-tooltip': {
        border: 'none',
        backgroundColor: tooltipBackground,
    },
    '.cm-tooltip .cm-tooltip-arrow:before': {
        borderTopColor: 'transparent',
        borderBottomColor: 'transparent',
    },
    '.cm-tooltip .cm-tooltip-arrow:after': {
        borderTopColor: tooltipBackground,
        borderBottomColor: tooltipBackground,
    },
    '.cm-tooltip-autocomplete': {
        '& > ul > li[aria-selected]': {
            backgroundColor: highlightBackground,
            color: fg,
        },
    },
}, { dark: true })

/// The highlighting style for code in the One Dark theme.
export const oneDarkHighlightStyle = HighlightStyle.define([
    { tag: t.namespace, color: yellow },
    { tag: t.keyword, color: purple },
    { tag: t.className, color: yellow },
    { tag: [t.variableName, t.operator], color: fg },
    { tag: [t.bool, t.null, t.typeName, t.number], color: orange },
    { tag: [t.function(t.variableName), t.function(t.propertyName)], color: blue },
    { tag: [t.propertyName, t.standard(t.function(t.variableName))], color: cyan },
    { tag: [t.local(t.variableName), t.standard(t.variableName)], color: red },
    { tag: t.comment, color: grey },
    { tag: t.string, color: green },
    { tag: [t.bracket, t.separator], color: light_grey },
])

/// Extension to enable the One Dark theme (both the editor theme and
/// the highlight style).
export const oneDark: Extension = [oneDarkTheme, syntaxHighlighting(oneDarkHighlightStyle)]
