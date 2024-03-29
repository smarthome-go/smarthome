<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import IconButton from '@smui/icon-button'
    import { onMount, tick } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar, data as userData, sleep } from '../../global'
    import Page from '../../Page.svelte'
    import Inputs from './Inputs.svelte'
    import AddHomescript from './dialogs/AddHomescript.svelte'
    import HmsSelector from './dialogs/HmsSelector.svelte'
    import { hmsLoaded, homescripts, jobs, loading } from './main'
    import { hmsEditorURLForId } from '../../urls'
    import DeleteHomescript from './dialogs/DeleteHomescript.svelte'
    import {
        killAllJobsById,
        type homescriptArgSubmit,
        type homescriptData,
        type homescriptResponseWrapper,
    } from '../../homescript'
    import { runHomescriptById, lintHomescriptById, getRunningJobs } from '../../homescript'
    import HmsArgumentPrompts from '../../components/Homescript/ArgumentPrompts/HmsArgumentPrompts.svelte'
    import ExecutionResultPopup from '../../components/Homescript/ExecutionResultPopup/ExecutionResultPopup.svelte'
    import TabBar from '@smui/tab-bar'
    import Tab from '@smui/tab'

    /*
        //// Dialog state management ////
     */
    let addOpen = false
    let deleteOpen = false

    let workspaces: string[] = []
    let workspace = 'default'

    // Is used when the run button is pressed and the current script has arguments
    let argumentsPromptOpen = false
    // Specifies whether the current argument prompts are used for linting or running
    let currentExecModeLint = false

    let selectedDataChanged = false
    let selection = ''
    let selectedData: homescriptData = {
        id: '',
        name: '',
        description: '',
        mdIcon: 'code',
        code: '',
        quickActionsEnabled: false,
        isWidget: false,
        schedulerEnabled: false,
        workspace: 'default',
    }

    // Using a copied `buffer` for the active script
    // Useful for a cancel feature
    $: if (selection !== '') updateSelectedData()

    // Updates the `selectedDataChanged` boolean
    // Which is used to disable the action buttons conditionally

    $: if (selectedData !== undefined && selection !== '') updateSelectedDataChanged()

    // Depending on whether the data has changed
    // the according boolean is updated
    function updateSelectedDataChanged() {
        const data = $homescripts.find(h => h.data.data.id === selection).data.data

        selectedDataChanged =
            data.name !== selectedData.name ||
            data.description !== selectedData.description ||
            data.mdIcon !== selectedData.mdIcon ||
            data.code !== selectedData.code ||
            data.schedulerEnabled !== selectedData.schedulerEnabled ||
            data.quickActionsEnabled !== selectedData.quickActionsEnabled ||
            data.isWidget !== selectedData.isWidget ||
            data.workspace !== selectedData.workspace
    }

    // Is used as soon as the active script is changed and is not empty
    function updateSelectedData() {
        const selectedDataTemp = $homescripts.find(h => h.data.data.id === selection).data.data

        // Static, contextual data
        selectedData.id = selectedDataTemp.id // Is required in order to send the request
        selectedData.code = selectedDataTemp.code // Required to preserve code

        // Changeable data
        selectedData.name = selectedDataTemp.name
        selectedData.description = selectedDataTemp.description
        selectedData.mdIcon = selectedDataTemp.mdIcon
        selectedData.quickActionsEnabled = selectedDataTemp.quickActionsEnabled
        selectedData.schedulerEnabled = selectedDataTemp.schedulerEnabled
        selectedData.isWidget = selectedDataTemp.isWidget
        selectedData.workspace = selectedDataTemp.workspace
    }

    // Is called when the changes have been successfully submitted and applied
    function updateSourceFromSelectedData() {
        // The index is required because JS clones the object
        const replaceIndex = $homescripts.findIndex(h => h.data.data.id === selection)
        $homescripts[replaceIndex].data.data.name = selectedData.name
        $homescripts[replaceIndex].data.data.description = selectedData.description
        $homescripts[replaceIndex].data.data.mdIcon = selectedData.mdIcon
        $homescripts[replaceIndex].data.data.quickActionsEnabled = selectedData.quickActionsEnabled
        $homescripts[replaceIndex].data.data.schedulerEnabled = selectedData.schedulerEnabled
        $homescripts[replaceIndex].data.data.isWidget = selectedData.isWidget
        $homescripts[replaceIndex].data.data.workspace = selectedData.workspace
        updateSelectedData()
    }

    // Fetches the available Homescripts for the selection and naming
    async function loadHomescripts() {
        $loading = true
        try {
            let res = await (await fetch('/api/homescript/list/personal/complete')).json()

            if (res.success !== undefined && !res.success) throw Error(res.error)
            homescripts.set(res) // Move the fetched homescripts into the store

            // Required because JS was created in 7 days
            await sleep(0)

            hmsLoaded.set(true) // Signal that the homescripts are loaded
        } catch (err) {
            $createSnackbar(`Could not load Homescripts: ${err}`)
        }
        $loading = false
    }

    // Requests modification of the currently-selected Homescript
    async function modifyCurrentHomescript() {
        $loading = true
        try {
            let res = await (
                await fetch('/api/homescript/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(selectedData),
                })
            ).json()
            if (!res.success) throw Error(res.error)

            // Create a new workspace if it does not exist already
            if (!workspaces.includes(selectedData.workspace))
                workspaces = [...workspaces, selectedData.workspace]

            const oldWorkspace = $homescripts.find(h => h.data.data.id === selectedData.id).data
                .data.workspace
            if ($homescripts.filter(h => h.data.data.workspace === oldWorkspace).length === 1)
                workspaces = workspaces.filter(w => w !== oldWorkspace)

            await tick()

            workspace = selectedData.workspace

            updateSourceFromSelectedData()
        } catch (err) {
            $createSnackbar(`Could not modify Homescript: ${err}`)
        }
        $loading = false
    }

    // Requests creation of a new Homescript
    async function createHomescript(data: homescriptData) {
        data.mdIcon = 'code'
        data.schedulerEnabled = false
        data.quickActionsEnabled = false
        $loading = true
        try {
            let res = await (
                await fetch('/api/homescript/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            // Append the new Homescript to the global store
            $homescripts = [
                ...$homescripts,
                {
                    arguments: [],
                    data: {
                        owner: $userData.userData.user.username,
                        data: data,
                    },
                },
            ]
            // The wait is required in order to delay the selection
            await sleep(50)

            // Create a new workspace if it does not exist already
            if (!workspaces.includes(data.workspace)) workspaces = [...workspaces, data.workspace]

            await tick()

            // Select the newly created workspace first
            workspace = data.workspace

            // Select the newly created Homescript for editing
            selection = data.id
            // Show the newly selected Homescript in the Inputs
            updateSelectedData()
        } catch (err) {
            $createSnackbar(`Could not create Homescript: ${err}`)
        }
        $loading = false
    }

    // Requests deletion of a Homescript
    async function deleteHomescript(id: string) {
        $loading = true
        try {
            let res = await (
                await fetch('/api/homescript/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            // Remove the current Homescript from the global store
            // An intermediate version of the list is required to prevent an edge case in which the script at index 0 is also the script to-be-deleted
            // In this case, a runtime error would occur in `Inputs.svelte`, and probably other places too
            const homescriptsTemp = $homescripts.filter(h => h.data.data.id !== selection)

            const wsToBeDeleted = $homescripts.find(h => h.data.data.id === id).data.data.workspace
            if ($homescripts.filter(h => h.data.data.workspace === wsToBeDeleted).length === 1)
                workspaces = workspaces.filter(w => w !== wsToBeDeleted)

            await tick()
            if (workspaces.length > 0) workspace = workspaces[0]

            // If no Homescript exist besides this one, only make changes persistent
            if (homescriptsTemp.length == 0) {
                selection = ''
                await sleep(10)
                $homescripts = homescriptsTemp
                $loading = false
            } else {
                // Select the first available Homescript as active
                selection = homescriptsTemp[0].data.data.id
                // Assign the intermediate list to the store in order to make changes persistent
                $homescripts = homescriptsTemp

                // Sleep 50ms in order to delay the selection update
                await sleep(50)

                // Show the newly selected Homescript in the Inputs
                updateSourceFromSelectedData()
            }
        } catch (err) {
            $createSnackbar(`Could not delete Homescript: ${err}`)
        }
        $loading = false
    }

    /*
       Executing the currently selected Homescript
    */
    let hmsExecutionResults: homescriptResponseWrapper[] = []

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    function initCurrentRun() {
        if ($homescripts.find(h => h.data.data.id === selection).arguments.length === 0) {
            runCurrentWithArgs([])
            setTimeout(async () => {
                $jobs = await getRunningJobs()
            }, 100)
            return
        }
        // The script is executed via callback: refer to the argument dialog
        currentExecModeLint = false
        argumentsPromptOpen = true
    }

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    function initCurrentLint() {
        if ($homescripts.find(h => h.data.data.id === selection).arguments.length === 0) {
            lintCurrentWithArgs([])
            return
        }
        // The script is linted via callback: refer to the argument dialog
        currentExecModeLint = true
        argumentsPromptOpen = true
    }

    // Used when the run button is pressed, error handling is accomplished here
    async function runCurrentWithArgs(args: homescriptArgSubmit[]) {
        try {
            const hmsRes = await runHomescriptById(selection, args, false)
            hmsExecutionResults = [
                ...hmsExecutionResults,
                {
                    response: hmsRes,
                    code: selectedData.code,
                    modeRun: true,
                },
            ]
            $jobs = await getRunningJobs()
        } catch (err) {
            $createSnackbar(`Could not execute ${selection}: ${err}`)
        }
    }

    // Used when the check button is pressed, error handling is accomplished here
    async function lintCurrentWithArgs(args: homescriptArgSubmit[]) {
        $loading = true
        try {
            const hmsRes = await lintHomescriptById(selection, args)
            hmsExecutionResults = [
                ...hmsExecutionResults,
                {
                    response: hmsRes,
                    code: selectedData.code,
                    modeRun: false,
                },
            ]
        } catch (err) {
            $createSnackbar(`Could not lint ${selection}: ${err}`)
        }
        $loading = false
    }

    // KEY BINDS.
    // CTRL + S => Save current script
    // F8       => Run current script
    // F9       => Lint current code
    // F10      => Cancel current job(s)
    document.addEventListener('keydown', e => {
        if (e.ctrlKey && e.key === 's') {
            e.preventDefault()

            if (!selectedDataChanged) {
                return
            }
            modifyCurrentHomescript()
        } else if (e.key === 'F8') {
            e.preventDefault()

            if ($jobs.filter(j => j.hmsId === selection).length > 0) {
                $createSnackbar('Program is already running.')
                return
            }

            if (selectedDataChanged) {
                $createSnackbar('Homescript has pending, unsaved changes.')
                return
            }

            initCurrentRun()
        } else if (e.key === 'F9') {
            e.preventDefault()

            if ($jobs.filter(j => j.hmsId === selection).length > 0) {
                $createSnackbar('Program is already running.')
                return
            }

            if (selectedDataChanged) {
                $createSnackbar('Homescript has pending, unsaved changes.')
                return
            }

            initCurrentLint()
        } else if (e.key === 'F10') {
            e.preventDefault()

            if ($jobs.filter(j => j.hmsId === selection).length === 0) {
                $createSnackbar('Homescript not running.')
                return
            }

            killCurrentJob()
        }
    })

    async function killCurrentJob() {
        await killAllJobsById(selection)
        await sleep(100)
        $jobs = await getRunningJobs()
    }

    onMount(async () => {
        // Load Homescripts as soon as the component is mounted
        await loadHomescripts()

        // Do some workspace-specific logic
        if ($homescripts.length > 0) {
            workspaces = [...new Set($homescripts.map(h => h.data.data.workspace))]
            workspace = workspaces[0]
            selection = $homescripts.find(h => h.data.data.workspace === workspace).data.data.id

            await tick()

            updateSelectedData()
        }

        // Load the currently running jobs
        $jobs = await getRunningJobs()
    })
</script>

<AddHomescript
    {workspace}
    on:add={event => {
        createHomescript(event.detail)
    }}
    bind:open={addOpen}
/>
<DeleteHomescript
    bind:data={selectedData}
    bind:open={deleteOpen}
    on:delete={event => deleteHomescript(event.detail.id)}
/>

{#if argumentsPromptOpen && $homescripts.find(h => h.data.data.id === selection) !== undefined && $homescripts.find(h => h.data.data.id === selection).arguments.length > 0}
    <HmsArgumentPrompts
        on:submit={event => {
            // Handle the decision between lint and run here
            if (currentExecModeLint) {
                lintCurrentWithArgs(event.detail)
            } else runCurrentWithArgs(event.detail)
        }}
        bind:open={argumentsPromptOpen}
        args={$homescripts.find(h => h.data.data.id === selection).arguments.map(a => a.data)}
    />
{/if}

{#if hmsExecutionResults[0] !== undefined}
    <ExecutionResultPopup
        open={true}
        data={hmsExecutionResults[0]}
        on:close={() => {
            // This hack is required so that the window still remains scrollable after removal
            setTimeout(() => (hmsExecutionResults = hmsExecutionResults.slice(1)), 1000)
        }}
    />
{/if}

<Page>
    <div id="header" class="mdc-elevation--z4">
        {#if workspaces.length > 0}
            <TabBar tabs={workspaces} bind:active={workspace} let:tab>
                <Tab {tab} minWidth>
                    <Label>{tab}</Label>
                </Tab>
            </TabBar>
        {:else}
            <h6>Homescript</h6>
        {/if}
        <div id="header__buttons">
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={async () => {
                    await loadHomescripts()
                    $jobs = await getRunningJobs()
                }}>refresh</IconButton
            >
            {#if $homescripts.length > 0}
                <Button on:click={() => (addOpen = true)}>
                    <Label>Script</Label>
                    <Icon class="material-icons">add</Icon>
                </Button>
            {/if}
        </div>
    </div>
    <Progress id="loader" bind:loading={$loading} />

    <div id="content">
        <div class="container mdc-elevation--z1" class:empty={$homescripts.length == 0}>
            <div class="homescripts" class:empty={$homescripts.length == 0 && $hmsLoaded}>
                {#if $homescripts.length == 0 && $hmsLoaded}
                    <i class="material-icons" id="no-homescripts-icon">code_off</i>
                    <h6 class="text-hint">No Homescripts</h6>
                    <Button on:click={() => (addOpen = true)} variant="outlined">
                        <Label>Create New</Label>
                        <Icon class="material-icons">add</Icon>
                    </Button>
                {:else}
                    <HmsSelector {workspace} bind:selection />
                {/if}
            </div>
        </div>
        <div id="inputs" class="mdc-elevation--z1" class:disabled={$homescripts.length == 0}>
            <div class="header">
                <h6>
                    Homescript {selection}
                </h6>
            </div>
            {#if $hmsLoaded && selection !== '' && selectedData !== undefined && $homescripts.find(h => h.data.data.id === selection) !== undefined}
                <Inputs bind:data={selectedData} bind:deleteOpen />
                <div class="run">
                    <div class="run__title">
                        <span class="text-hint">Code Actions</span>
                    </div>
                    <div class="run__buttons">
                        <Button
                            href={hmsEditorURLForId(selection)}
                            disabled={selectedDataChanged}
                            variant="outlined"
                        >
                            <Label>Code</Label>
                            <Icon class="material-icons">code</Icon>
                        </Button>
                        {#if $jobs.filter(j => j.hmsId === selection).length > 0}
                            <Button
                                on:click={killCurrentJob}
                                variant="outlined"
                            >
                                <Label>Kill</Label>
                                <Icon class="material-icons">cancel</Icon>
                            </Button>
                        {:else}
                            <Button
                                on:click={initCurrentRun}
                                disabled={selectedDataChanged}
                                variant="outlined"
                            >
                                <Label>Run</Label>
                                <Icon class="material-icons">play_arrow</Icon>
                            </Button>
                        {/if}
                        <Button
                            on:click={initCurrentLint}
                            disabled={selectedDataChanged}
                            variant="outlined"
                        >
                            <Label>Check</Label>
                            <Icon class="material-icons">bug_report</Icon>
                        </Button>
                    </div>
                </div>
                <div class="actions">
                    <Button on:click={updateSelectedData} disabled={!selectedDataChanged}>
                        <Label>Cancel</Label>
                    </Button>
                    <Button on:click={modifyCurrentHomescript} disabled={!selectedDataChanged}>
                        <Label>Apply Changes</Label>
                    </Button>
                </div>
            {/if}
        </div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding-right: 0.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;
        overflow-x: auto;

        h6 {
            margin: 0.5rem 0;
            margin-left: 1rem;
            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }

        &__buttons {
            display: flex;
            align-items: center;
        }
    }

    #no-homescripts-icon {
        font-size: 5rem;
        color: var(--clr-text-disabled);
    }

    .container {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        overflow: hidden;
        height: 100%;
        width: 50%;

        @include mobile {
            width: auto;
        }

        &.empty {
            width: 100%;
            background-color: transparent;
        }
    }

    #content {
        display: flex;
        flex-direction: column-reverse;
        margin: 1rem 1.5rem;
        gap: 1rem;
        transition-property: height;
        transition-duration: 0.3s;

        height: calc(100vh - 92px);
        flex-direction: row;

        @include mobile {
            flex-direction: column-reverse;
            height: 100%;
        }
    }

    .homescripts {
        height: 100%;
        overflow-y: auto;

        &.empty {
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 3rem;
            box-sizing: border-box;
            height: calc(100vh - 91px);
            width: 100%;
            gap: 1.5rem;

            h6 {
                margin: 0.5rem 0;
                font-size: 1.1rem;
            }

            @include mobile {
                gap: 1rem;
                height: calc(100vh - 143px);
                overflow: hidden;
            }
        }
    }
    #inputs {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        padding: 1.5rem;
        display: flex;
        flex-direction: column;
        width: 50%;

        h6 {
            margin: 0.5rem 0;
        }

        @include mobile {
            width: auto;
        }

        &.disabled {
            display: none;
        }
    }

    .actions {
        display: flex;
        justify-content: flex-end;
        gap: 0.5rem;
        margin-top: auto;

        @include mobile {
            margin-top: 1rem;
            flex-wrap: wrap;
        }
    }

    .run {
        margin-top: 2rem;
        background-color: var(--clr-height-1-3);
        border-radius: 0.4rem;
        padding: 0.9rem 1rem;

        @include mobile {
            background-color: transparent;
            border-radius: 0;
            padding: 0;
        }

        &__title {
            display: none;

            @include widescreen {
                display: block;
            }
        }

        &__buttons {
            display: flex;
            gap: 0.5rem;
            padding: 0.4rem 0;

            @include mobile {
                flex-direction: column;
            }
        }
    }
</style>
