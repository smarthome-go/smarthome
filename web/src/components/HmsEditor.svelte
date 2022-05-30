<script lang="ts">
    import {
        EditorState,
        EditorView,
        basicSetup,
    } from "@codemirror/basic-setup";
    import { EditorSelection, Compartment } from "@codemirror/state";
    import { javascript } from "@codemirror/lang-javascript";
    import { onMount } from "svelte";

    let editorDiv: HTMLElement = null;

    let myTheme = EditorView.theme(
        {
            "&": {
                color: "white",
                backgroundColor: "#034",
            },
            ".cm-content": {
                caretColor: "#0e9",
            },
            "&.cm-focused .cm-cursor": {
                borderLeftColor: "#0e9",
            },
            "&.cm-focused .cm-selectionBackground, ::selection": {
                backgroundColor: "#074",
            },
            ".cm-gutters": {
                backgroundColor: "#045",
                color: "#ddd",
                border: "none",
            },
        },
        { dark: true }
    );

    onMount(() => {
        let editor = new EditorView({
            state: EditorState.create({
                extensions: [basicSetup, javascript(), myTheme],
            }),
            parent: editorDiv,
        });

        editor.dispatch(
            editor.state.changeByRange((range) => ({
                changes: [
                    { from: range.from, insert: "_" },
                    { from: range.to, insert: "_" },
                ],
                range: EditorSelection.range(range.from + 2, range.to + 2),
            }))
        );
    });
</script>

<div id="editor" bind:this={editorDiv} />

<style lang="scss">
    #editor {
        width: 100%;
        height: 100%;
    }
</style>
