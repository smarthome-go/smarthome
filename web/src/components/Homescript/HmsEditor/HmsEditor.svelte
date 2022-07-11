<script lang="ts">
    import {
        EditorState,
        EditorView,
        basicSetup,
    } from "@codemirror/basic-setup";
    import { EditorSelection } from "@codemirror/state";
    import {indentWithTab} from '@codemirror/commands'
    import {keymap} from '@codemirror/view'
    import CodeMirror from "@codemirror/basic-setup";
    import { createEventDispatcher, onMount, setContext } from "svelte";
    //    import { tags } from "@lezer/highlight";
    // import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
    // import { completeFromList } from "@codemirror/autocomplete";

    // TODO: move local files to separate repository
    import { HomescriptLanguage, Homescript } from "./index.js";
    import { oneDark } from "./oneDark";

    const dispatch = createEventDispatcher()

    // Represents the editor's value
    export let code: string = "";
    $: setCode(code);

    function setCode(cd: string) {
        if (editor === undefined || editor.state.doc.toString() === cd) return;
        editor.dispatch(
            editor.state.update({
                changes: {from: 0, to: editor.state.doc.length, insert: cd}
            })
        );
    }

    // Will later be bound to the target of the CodeMirror editor
    let editorDiv: HTMLElement;

    let editor: EditorView;
    let timer: NodeJS.Timeout;

    onMount(() => {
        editor = new EditorView({
            state: EditorState.create({
                extensions: [
                    basicSetup,
                    keymap.of([indentWithTab]),
                    Homescript(),
                    oneDark,
                    EditorView.updateListener.of((v) => {
                        // TODO: lint / check code here
                        if (v.docChanged) {
                            if (timer) clearTimeout(timer);
                            timer = setTimeout(() => {
                                dispatch('update', code)
                            }, 500);
                        }
                    }),
                    EditorView.updateListener.of((v) => {
                        // TODO: lint / check code here
                        if (v.docChanged) {
                            code = editor.state.doc.toString();
                        }
                    }),
                ],
                doc: code,
            }),
            parent: editorDiv,
        });

        /*
        editor.dispatch(
            editor.state.changeByRange((range) => ({
                changes: [{ from: range.from, insert: "switch('id', on)" }],
                range: EditorSelection.range(range.from + 2, range.to + 2),
            }))
        );
         */
    });
</script>

<div class="hms-editor" bind:this={editorDiv} />

<style lang="scss">
    .hms-editor {
        height: 100%;
    }
</style>
