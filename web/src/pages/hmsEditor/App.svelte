<script lang="ts">
    import Page from '../../Page.svelte'
    import { createSnackbar } from '../../global'
    import { lintHomescriptCode } from '../../homescript'
    import type { homescript, homescriptArgSubmit, homescriptWithArgs } from '../../homescript'
    import { onMount } from 'svelte'
    import IconButton from '@smui/icon-button'
    import Select, { Option } from '@smui/select'
    import Progress from '../../components/Progress.svelte'

    // Custom HMS components
    import HmsEditor from '../../components/Homescript/HmsEditor/HmsEditor.svelte'
    import Terminal from '../../components/Homescript/ExecutionResultPopup/Terminal.svelte'
    import Button, { Label } from '@smui/button'
    import HmsArgumentPrompts from '../../components/Homescript/ArgumentPrompts/HmsArgumentPrompts.svelte'
    import type { hmsResWrapper } from './websocket'
    import Checkbox from '@smui/checkbox'
    import FormField from '@smui/form-field'
    import HmsFileExplorer from './HmsFileExplorer.svelte';
    import type { EditorHms } from './types';

    /*
       General variables
       Includes varialbes such as layout-management and loading indicators
     */
     // Whether the current script is a driver or a normal script

    // Specifies whether the argument prompt dialog should be open or closed
    let argumentsPromptOpen = false

    // Specifies whether the alternate layout (larger terminal) should be active or not
    let layoutAlt = false

    // Specifies if additional linter information should be shown
    // This raises the lint-level to `info`
    let showLintInfo = true

    // Is set to true when a script is linted or executed
    let requestLoading = false

    // Is set to true if either the script loads or is saved
    let otherLoading = false

    // If set to true, a banner (indicating that no script xyz has been found) is shown instead of the editor
    let err404 = false

    // Specifies the amount of jobs executing the current id (fetched initially)
    let currentExecutionCount = 0

    // Specifies the amount of jobs which the browser currently waits for
    // Used to limit the number of concurrent operations to exactly 1
    let currentExecutionHandles = 0

    /*
       Script management
       Variables and functions required to save all scripts
     */
    let homescripts: EditorHms[] = []

    // Is set to true as soon as the scripts are loaded
    // Required in the dynamic update of the current script (due to the list being empty when loaded=false)
    let homescriptsLoaded = false

    async function loadHomescript() {
        otherLoading = true
        try {
            const res = await (await fetch(`/api/homescript/list/personal/complete`)).json()
            if (res.success !== undefined && !res.success) throw Error(res.error)

            for (let script of res) {
                homescripts = [...homescripts, {
                    unsaved: false,
                    data: script,
                }]
            }

            homescriptsLoaded = true
            if (homescripts.length > 0) currentScript = homescripts[0].data.data.data.id

            // Populate the `savedCode` map so that for every script, there is already an entry in the map.
            // If this was not the case, changing a file will cause the UI to display an erronous `unsaved`.
            let savedCodeTemp: Map<string, string> = new Map()
            for (let script of homescripts) {
                savedCodeTemp.set(script.data.data.data.id, script.data.data.data.code)
            }
            savedCode = savedCodeTemp
        } catch (err) {
            $createSnackbar(`Failed to load editor for '${currentScript}': ${err}`)
        }
        otherLoading = false
    }

    // async function loadDrivers() {
    //     otherLoading = true
    //     try {
    //         const res = await (await fetch(`/api/system/hardware/drivers/list`)).json()
    //         if (res.success !== undefined && !res.success) throw Error(res.error)
    //
    //         for (let driver of res) {
    //             homescripts = [...homescripts, {
    //                 data: {
    //                     owner: "",
    //                     data: {
    //                         id: `${driver.vendorId}:${driver.modelId}`,
    //                         name: driver.name,
    //                         description: `Driver "${driver.name}"`,
    //                         mdIcon: "code", // TODO: change this to something nice
    //                         quickActionsEnabled: false,
    //                         isWidget: false,
    //                         code: driver.homescriptCode,
    //                         schedulerEnabled: false,
    //                         workspace: "drivers",
    //                     }
    //                 },
    //                 arguments: [],
    //             }]
    //         }
    //
    //         homescripts = res
    //         homescriptsLoaded = true
    //         if (homescripts.length > 0) currentScript = homescripts[0].data.data.id
    //     } catch (err) {
    //         $createSnackbar(`Failed to load editor for driver '${currentScript}': ${err}`)
    //     }
    //     otherLoading = false
    // }


    /*
       Current script management
       Saves which script is currently being edited
       Includes a function for changing the currently active script
     */
    // Saves the metadata of the current script (specified by URL query)
    let currentScript = ''

    let currentData: EditorHms = {
        unsaved: false,
        data: {
            data: {
                owner: '',
                data: {
                    id: currentScript,
                    name: '',
                    description: '',
                    mdIcon: '',
                    code: '',
                    quickActionsEnabled: false,
                    schedulerEnabled: false,
                    isWidget: false,
                    workspace: 'default',
                    type: 'NORMAL'
                },
            },
            arguments: [],
        }
    }

    $: if (currentData.data.data.data.code) {
        const savedCodeStr = savedCode.get(currentData.data.data.data.id)
        const currCodeStr = currentData.data.data.data.code
        const unsaved = savedCodeStr !== currCodeStr

        console.log(`saved=${savedCodeStr}, curr=${currCodeStr}`)

        if (unsaved != currentData.unsaved) {
            // This function produces a lot of runtime overhead, so be careful when calling it
            setUnsaved(unsaved)
        }
    }

    // Is called every time the `currentScript` variable changes
    $: if (homescriptsLoaded && currentScript) setCurrentScript()

    // Is used to update the currently shown script
    function setCurrentScript() {
        currentData = homescripts.find(h => h.data.data.data.id === currentScript)
        console.log(`set current data`, currentData)
        savedCode.set(currentData.data.data.data.id, currentData.data.data.data.code)
    }

    /*
       Code saving
       Variables and functions responsible for saving the code which is currently being edited
    */
    // Specifies whether there are unsaved changes or if the code is up-to-date
    let savedCode: Map<string, string> = new Map()

    // KEY BINDS
    // CTRL + S => Save current script
    // F8       => Run current script
    // F9       => Lint current code
    // F10      => Cancel current job(s)
    document.addEventListener('keydown', e => {
        if (e.ctrlKey && e.key === 's') {
            e.preventDefault()
            saveCurrent()
        } else if (e.key === 'F8') {
            if (currentExecutionHandles > 0) return
            if (currentData.unsaved) {
                $createSnackbar('This document contains unsaved changes')
                return
            }
            e.preventDefault()
            initCurrentRun()
        } else if (e.key === 'F9') {
            if (currentExecutionHandles > 0) return
            e.preventDefault()
            lintCurrentCode()
        } else if (e.key === 'F10') {
            e.preventDefault()
            killCurrentRun()
        }
    })

    // PREVENT the user from leaving the page if there are any unsaved changes
    window.addEventListener('beforeunload', e => {
        // Determine whether there are any unsaved changes
        let isAllSaved = true
        for (let script of homescripts) {
            if (script.unsaved) {
                isAllSaved = false
                break
            }
        }

        // If everything is saved, the user might safely navigate away from this page
        if (isAllSaved) {
            return
        }

        e.preventDefault();
        // https://developer.mozilla.org/en-US/docs/Web/API/Window/beforeunload_event
        // Included for legacy support, e.g. Chrome/Edge < 119
        e.returnValue = true;
    })

    // Sends a `save` request to the server, also updates the GUI display of unsaved changes to saved
    async function saveCurrent() {
        if (!currentData.unsaved) return
        otherLoading = true
        try {
            const res = await (
                await fetch(`/api/homescript/modify/code`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        id: currentData.data.data.data.id,
                        code: currentData.data.data.data.code,
                    }),
                })
            ).json()

            if (res.success !== undefined && !res.success) throw Error(res.error)

            savedCode.set(currentData.data.data.data.id, currentData.data.data.data.code)
            savedCode = savedCode
            // HACK: can this be implemented in a better way
            // this is required so that svelte triggers a refresh of the other variables

            // Commit this change
            setUnsaved(false)
        } catch (err) {
            $createSnackbar(`Failed to save '${currentScript}': ${err}`)
        }
        otherLoading = false
    }

    function setUnsaved(value: boolean) {
            currentData.unsaved = false

            // This part of the code is required to update the value in the list of homescripts.
            // It is required so that the file explorer also reports the correct save-state of its files.
            const index = homescripts.findIndex(h => h.data.data.data.id === currentData.data.data.data.id)
            if (index === -1) {
                throw "BUG warning: current homescript is not in the list of all homescripts"
            }

            homescripts[index].unsaved = value
    }

    /*
       Execution and linting
       Functions and variables responsible for linting and running the code
     */
    // Saves the last execution / lint result
    let currentExecRes: hmsResWrapper = undefined

    let output = ''
    $: if ((output || currentExecRes) && terminal) {
        terminal.scrollTop = terminal.scrollHeight
    }

    let terminal: HTMLDivElement

    // Keeps track of whether the current HMS request is meant to be `run` or `lint`
    // Is used in the argument-prompt popup which dispatches the according request to the server
    let currentExecModeLint = false

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    // Ported from `src/pages/homescript/App.svelte`
    function initCurrentRun() {
        if (homescripts.find(h => h.data.data.data.id === currentScript).data.arguments.length === 0) {
            runCurrentCode([])
            return
        }
        // The script is executed via callback: refer to the argument dialog
        currentExecModeLint = false
        argumentsPromptOpen = true
    }

    // Normal run functions
    async function runCurrentCode(args: homescriptArgSubmit[]) {
        try {
            output = ''
            currentExecRes = undefined
            if (currentData.data.data.data.code === '') {
                output = 'Nothing to run.'
            } else {
                runCodeWS(args)
            }
        } catch (err) {
            $createSnackbar(`Failed to run '${currentScript}': ${err}`)
        }
    }

    // Dry-run function without data modifications or expensive operations
    // Can be used to validate the correctness of a script without the need for execution
    async function lintCurrentCode() {
        currentExecModeLint = true
        requestLoading = true
        currentExecutionCount++
        currentExecutionHandles++
        try {
            const currentExecResTemp = await lintHomescriptCode(
                currentData.data.data.data.code,
                [],
                currentData.data.data.data.id,
                currentData.data.data.data.type == 'DRIVER',
            )
            let errs = currentExecResTemp.errors

            // If hint and info diagnostics should be hidden, do it here
            if (!showLintInfo)
                errs = errs.filter(d => {
                    if (d.diagnosticError !== null) {
                        if (d.diagnosticError.kind <= 1) {
                            return false
                        }
                    }
                    return true
                })

            currentExecRes = {
                code: currentData.data.data.data.code,
                modeRun: false,
                errors: errs,
                fileContents: currentExecResTemp.fileContents,
                success: currentExecResTemp.success,
            }
            output = currentExecResTemp.output
        } catch (err) {
            $createSnackbar(`Failed to lint '${currentScript}': ${err}`)
        }
        currentExecutionCount--
        currentExecutionHandles--
        requestLoading = false
    }

    async function killCurrentRun() {
        otherLoading = true
        try {
            if (conn !== undefined) {
                conn.send(JSON.stringify({ kind: 'kill' }))
            } else {
                console.error('This is bad')
            }
        } catch (err) {
            $createSnackbar(`Failed to terminate current run'${currentScript}': ${err}`)
        }
        otherLoading = false
    }

    let conn: WebSocket = undefined
    function runCodeWS(args: homescriptArgSubmit[]) {
        // Choose correct websocket protocol depending on the current context
        let protocol = undefined
        switch (document.location.protocol) {
            case 'http:':
                protocol = 'ws:'
                break
            case 'https:':
                protocol = 'wss:'
                break
            default:
                $createSnackbar(
                    `Unsupported protocol '${document.location.protocol}': only http and https are supported`,
                )
                return
        }

        // Build the websocket URL from the components
        let url = `${protocol}//${location.host}/api/homescript/run/ws`

        conn = new WebSocket(url)

        conn.onopen = () => {
            requestLoading = true
            currentExecutionCount++
            currentExecutionHandles++
            // Send the code to execute
            conn.send(JSON.stringify({ kind: 'init', payload: currentData.data.data.data.id, args }))
        }

        conn.onclose = () => {
            conn = undefined
            currentExecutionHandles--
            if (requestLoading) {
                $createSnackbar(`Websocket closed unexpectedly: connection lost`)
                output = 'Connection lost'
                currentExecutionCount--
                requestLoading = false
            }
        }

        conn.onmessage = evt => {
            try {
                let message = JSON.parse(evt.data)
                if (message.kind !== undefined && message.kind === 'out') output += message.payload
                else if (message.kind !== undefined && message.kind === 'res') {
                    currentExecRes = {
                        code: currentData.data.data.data.code,
                        modeRun: true,
                        errors: message.errors,
                        fileContents: message.fileContents,
                        success: message.success,
                    }
                    currentExecutionCount--
                    requestLoading = false
                } else if (message.kind === 'err') {
                    $createSnackbar(`Websocket error: ${message.message}`)
                    output = message.message
                    requestLoading = false
                    currentExecutionCount--
                } else {
                    console.error(message)
                    $createSnackbar(`Websocket error: unknown message kind: ${message.kind}`)
                }
            } catch (err) {
                console.error(err)
                $createSnackbar(`Websocket error: ${err}`)
            }
        }
    }

    // Load the Homescript-list at the beginning
    onMount(async () => {
        // Used for initially setting the active script via URL query
        const selectedFromQuery = new URLSearchParams(window.location.search).get('id')

        await loadHomescript()

        if (homescripts.find(h => h.data.data.data.id === selectedFromQuery) === undefined) {
            err404 = true
            return
        }

        currentScript = selectedFromQuery
    })
</script>

{#if argumentsPromptOpen && homescripts.find(h => h.data.data.data.id === currentScript).data.arguments.length > 0}
    <HmsArgumentPrompts
        on:submit={event => {
            // Handle the decision between lint and run here
            if (currentExecModeLint) {
                lintCurrentCode()
            } else runCurrentCode(event.detail)
        }}
        bind:open={argumentsPromptOpen}
        args={homescripts.find(h => h.data.data.data.id === currentScript).data.arguments.map(a => a.data)}
    />
{/if}

<Page persistentSlimNav={true}>
    {#if err404}
        <div id="error404">
            <i class="material-icons" id="no-automations-icon">edit_off</i>
            <h6 class="text-hint">Not found</h6>
            <Button on:click={() => (window.location.href = '/homescript')}>
                <Label>View Homescripts</Label>
            </Button>
        </div>
    {:else}
        <div id="header" class="mdc-elevation--z4">
            <div id="header__left">
                <span>Editing {currentData.data.data.data.name} </span>
                <div id="header__left__save" class:unsaved={currentData.unsaved}>
                    <i class="material-icons"
                        >{currentData.unsaved ? 'backup' : 'cloud_done'}</i
                    >
                    {currentData.unsaved ? 'unsaved' : 'saved'}
                </div>
                {#if currentExecRes !== undefined}
                    <div
                        id="header__left__errors"
                        class:error={!currentExecRes.modeRun && !currentExecRes.success}
                    >
                        <i class="material-icons">{currentExecRes.success ? 'done' : 'error'}</i>
                        {currentExecRes.success ? 'working' : 'errors'}
                    </div>
                {/if}
            </div>
            <div id="header__buttons">
                <IconButton class="material-icons" on:click={() => (layoutAlt = !layoutAlt)}
                    >vertical_split</IconButton
                >
                <IconButton
                    class="material-icons"
                    on:click={saveCurrent}
                    disabled={currentData.unsaved}>save</IconButton
                >
                <Progress type="circular" bind:loading={otherLoading} />
            </div>
        </div>
        <div class="container">
            <div class="container__left">
                <div class="container__left__files">
                    <span class="text-hint mdc-elevation--z2 container__left__files__title">Files</span>
                    <HmsFileExplorer
                        bind:homescripts
                        bind:currentScript={currentData}
                    ></HmsFileExplorer>
                </div>

                <div class="container__left__diagnostics">
                    <div class="container__left__diagnostics__list">
                        {#if currentExecRes !== undefined}
                            <div class="container__left__diagnostics__list__item">
                                <span class='icon-info'></span>
                                <span>
                                {currentExecRes.errors.map(e => e.diagnosticError !== null ? (e.diagnosticError.kind === 1 ? 1 : 0) : 0).reduce((acc, i) => acc + i, 0)}
                                </span>
                            </div>
                            <div class="container__left__diagnostics__list__item">
                                <span class='icon-warn'></span>
                                <span>
                                {currentExecRes.errors.map(e => e.diagnosticError !== null ? (e.diagnosticError.kind === 2 ? 1 : 0) : 0).reduce((acc, i) => acc + i, 0)}
                                </span>
                            </div>
                            <div class="container__left__diagnostics__list__item">
                                <span class='icon-error'></span>
                                <span>
                                {currentExecRes.errors.map(e => e.diagnosticError !== null ? (e.diagnosticError.kind === 3 ? 1 : 0) : (e.syntaxError !== null ? 1 : 0)).reduce((acc, i) => acc + i, 0)}
                                </span>
                            </div>
                        {:else}
                            <span class='text-disabled'>
                                No diagnostics available
                            </span>
                        {/if}
                    </div>
                </div>
            </div>
            <div class="container__editor" class:alt={layoutAlt}>
                {#if homescriptsLoaded}
                <HmsEditor
                    bind:moduleName={currentData.data.data.data.id}
                    bind:code={currentData.data.data.data.code}
                    {showLintInfo}
                    isDriver={currentData.data.data.data.type === 'DRIVER'}
                />
                {/if}
            </div>
            <div class="container__terminal" class:alt={layoutAlt}>
                <div class="container__terminal__header mdc-elevation--z2">
                    <div class="container__terminal__header__left">
                        <IconButton
                            class="material-icons"
                            on:click={initCurrentRun}
                            disabled={currentData.unsaved ||
                                currentExecutionHandles > 0}>play_arrow</IconButton
                        >
                        <IconButton
                            class="material-icons"
                            on:click={lintCurrentCode}
                            disabled={currentExecutionHandles > 0}
                        >
                            bug_report</IconButton
                        >
                        <IconButton
                            class="material-icons"
                            on:click={killCurrentRun}
                            disabled={currentExecutionCount === 0}
                        >
                            cancel</IconButton
                        >
                        <IconButton
                            class="material-icons"
                            on:click={() => {
                                currentExecRes = undefined
                                output = ''
                            }}
                            disabled={requestLoading ||
                                (currentExecRes == undefined && output == '')}>replay</IconButton
                        >
                    </div>
                    <div class="container__terminal__header__right">
                        <FormField>
                            <Checkbox bind:checked={showLintInfo} />
                            <span slot="label">show info</span>
                        </FormField>
                    </div>
                </div>
                <Progress type="linear" bind:loading={requestLoading} />
                <div class="container__terminal__content" bind:this={terminal}>
                    {#if output.length === 0 && currentExecRes === undefined}
                        <span class="gray"> Homescript output will be displayed here. </span>
                    {:else}
                        <Terminal data={currentExecRes} {output} />
                    {/if}
                </div>
            </div>
        </div>
    {/if}
</Page>

<style lang="scss">
    @use '../../mixins' as *;
    @use '../../components/Homescript/HmsEditor/icons.scss' as *;

    #error404 {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-top: 8.5rem;
        gap: 1rem;

        i {
            font-size: 5rem;
            color: var(--clr-text-disabled);
        }

        h6 {
            margin: 0.5rem 0;
        }
    }

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        height: 3.5rem;

        @include mobile {
            // Hides bottom line
            height: 3.4rem;
            justify-content: flex-start;
        }

        &__left {
            display: flex;
            gap: 1rem;

            span {
                @include mobile {
                    display: none;
                }
            }

            &__save,
            &__errors {
                color: var(--clr-text-disabled);
                display: flex;
                align-items: center;
                gap: 0.4rem;
                font-size: 0.9rem;

                @include mobile {
                    display: none;
                }

                i {
                    font-size: 1.25em;
                }

                &.unsaved,
                &.error {
                    color: var(--clr-error);
                }
            }
        }

        &__buttons {
            display: flex;
            align-items: center;
            gap: 0.2rem;
        }
    }

    .container {
        display: flex;
        overflow: hidden;
        flex-direction: column;
        height: calc(100vh - 3.67rem);

        @include mobile {
            height: calc(100vh - 9rem);
        }

        @include widescreen {
            flex-direction: row;
        }

        &__left {
            display: flex;
            flex-direction: column;

            &__files {
                width: 20rem;
                display: flex;
                flex-direction: column;

                &__title {
                    padding: .3rem .8rem;
                }
            }

            &__diagnostics {
                margin-top: auto;

                &__list {
                    display: flex;
                    gap: .6rem;
                    background-color: var(--clr-height-1-4);
                    padding: .2rem .5rem;

                    &__item {
                        display: flex;
                        align-items: center;
                        gap: .2rem;
                        color: var(--clr-text-hint);

                        .icon-info {
                           content: url($info-icon-svg);
                           height: 1rem;
                        }

                        .icon-warn {
                           content: url($warn-icon-svg);
                           height: 1rem;
                        }

                        .icon-error {
                           content: url($error-icon-svg);
                           height: 1rem;
                        }
                    }
                }
            }
        }

        &__editor {
            overflow: auto;
            height: 75%;

            @include widescreen {
                width: 75%;
                height: 100%;
            }

            // Used when the expand-terminal button is selected
            transition-property: width, height;
            transition-duration: 0.25s;

            &.alt {
                @include widescreen {
                    width: 25%;
                }
                @include not-widescreen {
                    height: 25%;
                }
            }
        }

        &__terminal {
            height: 25%;
            display: flex;
            flex-direction: column;

            // Used when the expand-terminal button is selected
            transition-property: width, height;
            transition-duration: 0.25s;

            @include widescreen {
                width: 25%;
                height: 100%;
            }

            &.alt {
                @include widescreen {
                    width: 75%;
                }
                @include not-widescreen {
                    height: 75%;
                }
            }

            &__header {
                padding: 0.2rem;
                background-color: var(--clr-height-0-1);
                display: flex;
                align-items: center;
                justify-content: space-between;

                &__right {
                    padding-right: 1rem;
                }
            }

            &__content {
                font-family: 'Jetbrains Mono', monospace;
                font-size: 0.9rem;
                padding: 1rem 1.3rem;
                height: 100%;
                overflow-y: auto;
            }
        }
    }
</style>
