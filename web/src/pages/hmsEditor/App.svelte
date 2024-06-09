<script lang="ts">
    import Split, { type SplitInstance } from 'split-grid'
    import Page from '../../Page.svelte'
    import { createSnackbar } from '../../global'
    import { lintHomescriptCode } from '../../homescript'
    import type { homescript, homescriptArgSubmit, homescriptResponse, homescriptWithArgs } from '../../homescript'
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
    import EditorLeft from './EditorLeft.svelte';

    /*
       General variables
       Includes variables such as layout-management and loading indicators
     */
     // Whether the current script is a driver or a normal script

    // Specifies whether the argument prompt dialog should be open or closed
    let argumentsPromptOpen = false

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
                    errors: false,
                    data: script,
                }]
            }

            homescriptsLoaded = true
            if (homescripts.length > 0) currentScript = homescripts[0].data.data.data.id

            // Populate the `savedCode` map so that for every script, there is already an entry in the map.
            // If this was not the case, changing a file will cause the UI to display an erroneous `unsaved`.
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
        errors: false,
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
            setBooleanPropertyOnHms('unsaved', unsaved)
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

    // KEY BINDS.
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

    let lastSaveHadError = false

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
            setBooleanPropertyOnHms('unsaved', false)


            if (lastSaveHadError) {
                lastSaveHadError = false
                $createSnackbar("Saved successfully")
            }
        } catch (err) {
            $createSnackbar(`Failed to save '${currentScript}': ${err}`)
            lastSaveHadError = true
        }
        otherLoading = false
    }

    function setBooleanPropertyOnHms(propertyKey: 'unsaved' | 'errors', value: boolean) {
            currentData[propertyKey] = value

            // This part of the code is required to update the value in the list of homescripts.
            // It is required so that the file explorer also reports the correct save-state of its files.
            const index = homescripts.findIndex(h => h.data.data.data.id === currentData.data.data.data.id)
            if (index === -1) {
                throw "BUG warning: current homescript is not in the list of all homescripts"
            }

            homescripts[index][propertyKey] = value
    }

    /*
       Execution and linting
       Functions and variables responsible for linting and running the code
     */
    // Saves the last execution / lint result
    let currentExecRes: hmsResWrapper = undefined

    let output = ''
    $: if ((output || currentExecRes) && terminalContentDiv) {
        terminalContentDiv.scrollTop = terminalContentDiv.scrollHeight
    }

    let terminalContentDiv: HTMLDivElement

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

    function displayLintResult(res: homescriptResponse) {
        let errs = res.errors

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

        let fileContents: Map<string, string> = new Map();
        for (let key of Object.keys(res.fileContents)) {
        fileContents.set(key, res.fileContents[key])
        }

        res.fileContents

        currentExecRes = {
            // code: currentData.data.data.data.code,
            modeRun: false,
            errors: errs,
            fileContents,
            success: res.success,
        }

        setBooleanPropertyOnHms('errors', !res.success)

        output = res.output
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

            displayLintResult(currentExecResTemp)
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
            conn.send(JSON.stringify({
                kind: 'init',
                hmsID: currentData.data.data.data.id,
                args,
            }))
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

        let WSBufferSizeChars = 1

        let WSOutputBuf = ""
        conn.onmessage = evt => {
            try {
                let message = JSON.parse(evt.data)
                if (message.kind !== undefined && message.kind === 'out') {
                    WSOutputBuf += message.payload
                    if (WSOutputBuf.length >= WSBufferSizeChars) {
                        output += WSOutputBuf;
                        WSOutputBuf = ""
                    }
                }
                else if (message.kind !== undefined && message.kind === 'res') {
                    let fileContents: Map<string, string> = new Map();

                    for (let key of Object.keys(message.fileContents)) {
                        fileContents.set(key, message.fileContents[key])
                    }

                    currentExecRes = {
                        modeRun: true,
                        errors: message.errors,
                        fileContents,
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

    //
    // Resizing logic.
    //

    let terminalDiv: HTMLElement | undefined = undefined

    let terminalResizing = false
    let editorLeftResizing = false

    const gutterLeftTrackNr = 1
    const gutterTerminalTrackNr = 3

    function updateTerminalSizeAfterDrag(_direction: 'row' | 'column', track:number) {
        switch (track) {
            case gutterLeftTrackNr:
                editorLeftResizing = false
            case gutterTerminalTrackNr:
                const terminalContentStyle =window.getComputedStyle(terminalContentDiv)
                const terminalContentPaddingHorizontal = parseFloat(terminalContentStyle.paddingLeft)
                    + parseFloat(terminalContentStyle.paddingRight)
                const parentWidth = terminalDiv.getBoundingClientRect().width
                const maxWidth = parentWidth - terminalContentPaddingHorizontal
                terminalContentDiv.style.maxWidth = `${maxWidth}px`
                terminalResizing = false
                break
        default:
        }
    }

    function blurTerminalOnDrag(_direction: 'row' | 'column', track: number) {
        switch (track) {
            case gutterLeftTrackNr:
                editorLeftResizing = true
                break
            case gutterTerminalTrackNr:
                terminalResizing = true
                break
            default:
        }
    }


    let split: SplitInstance = null

    // Load the Homescript-list at the beginning
    onMount(async () => {
        let onMobile = window.matchMedia("(max-width: 47rem)")
        onMobile.addEventListener("change", () => {
            registerMobile(onMobile)
        })

        // Used for initially setting the active script via URL query
        const selectedFromQuery = new URLSearchParams(window.location.search).get('id')

        await loadHomescript()

        if (homescripts.find(h => h.data.data.data.id === selectedFromQuery) === undefined) {
            err404 = true
            return
        }

        currentScript = selectedFromQuery

        registerMobile(onMobile)
    })

    function registerDesktop(mediaQuery: MediaQueryList) {
        if (mediaQuery.matches) {
            return
        }

        if (split)
            split.destroy()

        console.log("Register desktop...")

        // Resizing setup.
        split = Split({
            onDragEnd: updateTerminalSizeAfterDrag,
            onDragStart: blurTerminalOnDrag,

            columnMinSizes: { 0: 50, 2: 50, 4: 50 },
            columnGutters: [
            {
                track: 1,
                element: document.querySelector('.container__gutter-1'),
            },
            {
                track: 3,
                element: document.querySelector('.container__gutter-3'),
            }
            ],
        })

        updateTerminalSizeAfterDrag('column', 3)
    }

    let container: HTMLDivElement = null
    const GUTTER_WIDTH = 5;

    function registerMobile(mediaQuery: MediaQueryList) {
        // TODO: solve this!

        if (!mediaQuery.matches) {
            registerDesktop(mediaQuery)
            return
        }

        if (split)
            split.destroy()


        // container.style.gridTemplateRows = `2.5fr ${GUTTER_WIDTH} 9fr ${GUTTER_WIDTH} 350px`

        console.log("Register mobile...")

        split = Split({
            rowMinSizes: { 0: 30, 2: 50, 4: 55 },
            gridTemplateRows: `1fr 8px 4fr 8px 1fr`,

            onDragStart: (direction, track) => {
            },
            onDragEnd: (direction, track) => {
            },

            rowGutters: [
            {
                track: 1,
                element: document.querySelector('.container__gutter-1'),
            },
            {
                track: 3,
                element: document.querySelector('.container__gutter-3'),
            }
            ],
        })
    }
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
                <IconButton
                    class="material-icons"
                    on:click={saveCurrent}
                    disabled={currentData.unsaved}>save</IconButton
                >
                <Progress type="circular" bind:loading={otherLoading} />
            </div>
        </div>

        <div class="container" bind:this={container}>
            <div class="container__left" class:resizing={editorLeftResizing}>
                <EditorLeft
                    bind:currentData
                    bind:homescripts
                    {currentExecRes}
                />
            </div>

            <div class="container__gutter-col container__gutter-1"></div>

            <div class="container__editor">
                {#if homescriptsLoaded}
                <HmsEditor
                    on:lint={(e) => displayLintResult(e.detail)}
                    bind:moduleName={currentData.data.data.data.id}
                    bind:code={currentData.data.data.data.code}
                    {showLintInfo}
                    isDriver={currentData.data.data.data.type === 'DRIVER'}
                />
                {/if}
            </div>

            <div class="container__gutter-col container__gutter-3"></div>

            <div class="container__terminal" bind:this={terminalDiv} class:resizing={terminalResizing}>
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
                <div class="container__terminal__content" bind:this={terminalContentDiv}>
                    <!-- {#if terminalResizing} -->
                    <!--     Resizing... -->
                    <!-- {:else} -->
                        {#if output.length === 0 && currentExecRes === undefined}
                            <span class="gray"> Homescript output will be displayed here. </span>
                        {:else}
                            <Terminal data={currentExecRes} {output} />
                        {/if}
                    <!-- {/if} -->
                </div>
            </div>
        </div>
    {/if}
</Page>

<style lang="scss">
    @use '../../mixins' as *;

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
        $gutter-width: 5px;
        $gutter-color0: var(--clr-height-0-3);
        $gutter-color1: var(--clr-height-4-8);

        display: grid;
        grid-template-columns: 2.5fr $gutter-width 9fr $gutter-width 350px;

        @include mobile {
            grid-template-rows: 2.5fr $gutter-width 9fr $gutter-width 350px;
            grid-template-columns: unset;
            height: calc(100vh - 6.9rem);
        }

        height: calc(100vh - 3.67rem);
        //overflow: hidden;


        &__gutter-col {
            grid-row: 1/-1;

            cursor: col-resize;
            background: repeating-linear-gradient(
                45deg,
                $gutter-color0 0px,
                $gutter-color0 2px,
                $gutter-color1 2px,
                $gutter-color1 $gutter-width
            );
        }

        &__gutter-1 {
            grid-column: 2;

            @include mobile {
                grid-row: 2;
                grid-column: unset;
            }
        }

        &__gutter-3 {
            grid-column: 4;

            @include mobile {
                grid-row: 4;
                grid-column: unset;
            }
        }

        @mixin on-resize {
            &.resizing {
                background: repeating-linear-gradient(
                    45deg,
                    var(--clr-height-0-1) 0px,
                    var(--clr-height-0-1) 5px,
                    transparent 5px,
                    transparent 10px
                );
            }
        }

        &__left {
            display: flex;
            flex-direction: column;

            @include mobile {
                overflow-y: auto;
            }

            @include on-resize;
        }

        &__editor {
            overflow: auto;
            height: 100%;
        }

        &__terminal {
            height: 100%;
            display: flex;
            flex-direction: column;
            overflow-x: hidden;

            @include on-resize;

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
                font-family: 'Jetbrains Mono NL', monospace;
                font-size: 0.9rem;
                padding: 1rem 1.3rem;
                height: 100%;
                overflow-y: auto;

                // This is updated via JS after a drag.
                // Required because otherwise, a lot of output would grow the terminal.
                // However, the terminal should only grow when the user drags it.
                max-width: 10rem;
            }
        }
    }
</style>
