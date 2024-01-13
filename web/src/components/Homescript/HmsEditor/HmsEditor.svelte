<script lang="ts">
    import { EditorView, basicSetup } from 'codemirror'
    import { EditorState } from '@codemirror/state'
    import { indentWithTab } from '@codemirror/commands'
    import { keymap, drawSelection, dropCursor } from '@codemirror/view'
    import { linter, lintGutter, type Diagnostic } from '@codemirror/lint'
    import { createEventDispatcher, onMount } from 'svelte'
    import { indentUnit } from '@codemirror/language'
    import { Homescript } from 'codemirror-lang-homescript'
    // import { Homescript } from './index'
    import { oneDark } from './oneDark'
    import { lintHomescriptCode } from '../../../homescript'
    import { createSnackbar } from '../../../global'

    const dispatch = createEventDispatcher()

    // Specifies whether the editor should register a CTRL+S catcher
    // This catcher is intended to prevent the browser's default action
    // However, the catcher also emits a change event when the key combination is pressed
    export let registerCtrlSCatcher = false

    // Specifies whether this program is a hardware device driver or a normal program
    export let isDriver = false

    // Represents the editor's value
    export let code = ''
    $: setCode(code)

    // Can be bound to give the code a module name
    export let moduleName = ''

    // Whether the editor should show diagnostics with the `info` level
    export let showLintInfo = true
    $: if (showLintInfo !== undefined) triggerUpdate()

    function triggerUpdate() {
        // Reload the diagnostics if this value changes
        if (editor !== undefined) {
            let oldCode = code
            // Updates the code so that new diagnostics can be seen
            setCode((code += ' '))
            setCode(oldCode)
        }
    }

    function setCode(cd: string) {
        if (editor === undefined || editor.state.doc.toString() === cd) return
        editor.dispatch(
            editor.state.update({
                changes: { from: 0, to: editor.state.doc.length, insert: cd },
            }),
        )
    }

    // Will later be bound to the target of the CodeMirror editor
    let editorDiv: HTMLElement

    let editor: EditorView

    // eslint-disable-next-line no-undef
    let timer: NodeJS.Timeout

    // TODO: check filenames + syntaax erorrs in imported module

    const HMSlinter = linter(async () => {
        let diagnostics: Diagnostic[] = []

        try {
            console.log('linting as driver: ', isDriver)
            const result = await lintHomescriptCode(code, [], moduleName, isDriver)
            diagnostics = result.errors.map(e => {
                let severity = 'error'
                let message = 'error: unknown'
                let kind = 'error: unknown'

                let notes = []

                // everything except diagnostics will be a standard `error`
                if (e.syntaxError !== null) {
                    message = e.syntaxError.message
                    kind = 'SyntaxError'
                } else if (e.diagnosticError !== null) {
                    message = e.diagnosticError.message
                    switch (e.diagnosticError.kind) {
                        case 0:
                            kind = 'Hint'
                            severity = 'info'
                            break
                        case 1:
                            kind = 'Info'
                            severity = 'info'
                            break
                        case 2:
                            kind = 'Warning'
                            severity = 'warning'
                            break
                        case 3:
                            kind = 'Error'
                            severity = 'error'
                            break
                    }

                    for (let note of e.diagnosticError.notes) {
                        notes.push(`- note: ${note}`)
                    }
                } else if (e.runtimeError) {
                    throw 'A runtime error cannot occur during analysis'
                }

                return Object.create({
                    from: e.span.start.index,
                    to:
                        e.span.end.index + 1 <= code.length
                            ? e.span.end.index + 1
                            : e.span.end.index,
                    severity: severity,
                    message: `${kind}: ${message}\n${notes.join('\n')}`,
                    source: 'Homescript analyzer',
                })
            })
        } catch (err) {
            $createSnackbar(`Failed to lint: ${err}`)
            console.error(err)
        }

        if (showLintInfo) return diagnostics
        else return diagnostics.filter(d => d.severity !== 'info')
    })

    onMount(() => {
        if (registerCtrlSCatcher)
            document.addEventListener('keydown', e => {
                if (e.ctrlKey && e.key === 's') {
                    e.preventDefault()
                    dispatch('update', code)
                }
            })

        editor = new EditorView({
            state: EditorState.create({
                extensions: [
                    basicSetup,
                    drawSelection(),
                    dropCursor(),
                    indentUnit.of('    '),
                    keymap.of([indentWithTab]),
                    Homescript(),
                    oneDark,
                    // Linting
                    HMSlinter,
                    lintGutter(),
                    // Emit the `update` event 500 ms after the last keystroke
                    EditorView.updateListener.of(v => {
                        if (v.docChanged) {
                            if (timer) clearTimeout(timer)
                            timer = setTimeout(() => {
                                dispatch('update', code)
                            }, 500)
                        }
                    }),
                    // Update the component code on every change
                    EditorView.updateListener.of(v => {
                        if (v.docChanged) {
                            code = editor.state.doc.toString()
                        }
                    }),
                ],
                doc: code,
            }),
            parent: editorDiv,
        })
    })
</script>

<div class="hms-editor" bind:this={editorDiv} />

<style lang="scss">
    @use '../../../components/Homescript/HmsEditor/icons.scss' as *;

    .hms-editor {
        height: 100%;
    }

    :global {
        .cm-lint-marker-info {
            content: url($info-icon-svg) !important;
        }
        .cm-lint-marker-warning {
            content: url($warn-icon-svg) !important;
        }
        .cm-lint-marker-error {
            content: url($error-icon-svg) !important;
        }

        .ͼ4 .cm-line ::selection {
            background-color: rgba(255, 255, 255, 0.2) !important;
        }

        .cm-selectionMatch {
            background-color: rgba(100, 255, 0, 0.05) !important;
        }
    }
</style>
