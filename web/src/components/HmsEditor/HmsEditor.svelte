<script lang="ts">
    import {
        EditorState,
        EditorView,
        basicSetup,
    } from "@codemirror/basic-setup";
    import { EditorSelection, Compartment } from "@codemirror/state";
    import { onMount } from "svelte";
    import { tags } from "@lezer/highlight";
    import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
    import { completeFromList } from "@codemirror/autocomplete";
    // TODO: move local files to separate repository
    import { HomescriptLanguage, Homescript } from "./index.js";
    import { oneDark } from "./oneDark";

    const myHighlightStyle = HighlightStyle.define([
        { tag: tags.keyword, color: "#fc6" },
        { tag: tags.number, color: "#fc6" },
        { tag: tags.comment, color: "#f5d", fontStyle: "italic" },
    ]);

    let editorDiv: HTMLElement = null;

    onMount(() => {
        let editor = new EditorView({
            state: EditorState.create({
                extensions: [basicSetup, Homescript(), oneDark],
            }),
            parent: editorDiv,
        });

        editor.dispatch(
            editor.state.changeByRange((range) => ({
                changes: [{ from: range.from, insert: "switch('id', on)" }],
                range: EditorSelection.range(range.from + 2, range.to + 2),
            }))
        );
    });
</script>

<div id="editor" bind:this={editorDiv} />

<style lang="scss">
    #editor {
        height: 100rem;
    }
</style>
