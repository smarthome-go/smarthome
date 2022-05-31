<script lang="ts">
    import {
        EditorState,
        EditorView,
        basicSetup,
    } from "@codemirror/basic-setup";
    import { EditorSelection } from "@codemirror/state";
    import { onMount } from "svelte";
    //    import { tags } from "@lezer/highlight";
    // import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
    // import { completeFromList } from "@codemirror/autocomplete";

    // TODO: move local files to separate repository
    import { HomescriptLanguage, Homescript } from "./index.js";
    import { oneDark } from "./oneDark";

    // Will later be binded to the target of the CodeMirror editor
    let editorDiv: HTMLElement;

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

<div class="hms-editor" bind:this={editorDiv} />

<style lang="scss">
    .hms-editor {
        height: 100rem;
    }
</style>
